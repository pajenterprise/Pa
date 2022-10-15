import json
import os
import subprocess
from pprint import pprint

from .common.gitlab import Gitlab


def check_failed_status(project_name, pipeline_id):
    gitlab = Gitlab()

    jobs = gitlab.all_jobs(project_name, pipeline_id)

    # Get instances of failed jobs
    failed_jobs = {job["name"]: [] for job in jobs if job["status"] == "failed"}

    # Group jobs per name
    for job in jobs:
        if job["name"] in failed_jobs:
            failed_jobs[job["name"]].append(job)

    # There, we now have the following map:
    # job name -> list of jobs with that name, including at least one failed job

    final_failed_jobs = []
    for job_name, jobs in failed_jobs.items():
        # We sort each list per creation date
        jobs.sort(key=lambda x: x["created_at"])
        # Check the final job in the list: it contains the current status of the job
        final_status = {
            "name": job_name,
            "id": jobs[-1]["id"],
            "stage": jobs[-1]["stage"],
            "status": jobs[-1]["status"],
            "allow_failure": jobs[-1]["allow_failure"],
            "url": jobs[-1]["web_url"],
            "retry_summary": [job["status"] for job in jobs],
        }
        final_failed_jobs.append(final_status)

    pprint(final_failed_jobs)

    return final_failed_jobs


class Test:
    PACKAGE_PREFIX = "github.com/DataDog/datadog-agent/"

    def __init__(self, owners, name, package):
        self.name = name
        self.package = self.__removeprefix(package)
        self.owners = self.__get_owners(owners, package)

    def __removeprefix(self, package):
        return package[len(self.PACKAGE_PREFIX) :]

    def __get_owners(self, OWNERS, package):
        owners = OWNERS.of(self.__removeprefix(package))
        return [name for (kind, name) in owners if kind == "TEAM"]


def get_failed_tests(project_name, job):
    gitlab = Gitlab()

    # Get codeowners
    from codeowners import CodeOwners

    CODEOWNERS_PATH = os.path.join(os.path.dirname(os.path.realpath(__file__)), "..", "..", ".github", "CODEOWNERS")
    with open(CODEOWNERS_PATH, 'r') as f:
        codeowners = f.read()
    OWNERS = CodeOwners(codeowners)

    test_output = gitlab.artifact(project_name, job["id"])
    for line in test_output.splitlines():
        try:
            json_test = json.loads(line)
            if "message" in json_test:
                continue
            if (
                'Test' in json_test
                and 'Package' in json_test
                and 'Action' in json_test
                and json_test["Action"] == "fail"
            ):
                yield Test(OWNERS, json_test['Test'], json_test['Package'])
        except Exception as e:
            print("WARN: parsing", line, "failed", e)


def prepare_global_failure_message(header, failed_jobs):
    message = """{header} pipeline <{pipeline_url}|{pipeline_id}> for {commit_ref_name} failed.
{commit_title} (<{commit_url}|{commit_short_sha}>) by {author}
Failed jobs:""".format(
        header=header,
        pipeline_url=os.getenv("CI_PIPELINE_URL"),
        pipeline_id=os.getenv("CI_PIPELINE_ID"),
        commit_ref_name=os.getenv("CI_COMMIT_REF_NAME"),
        commit_title=os.getenv("CI_COMMIT_TITLE"),
        commit_url="{project_url}/commit/{commit_sha}".format(
            project_url=os.getenv("CI_PROJECT_URL"), commit_sha=os.getenv("CI_COMMIT_SHA")
        ),
        commit_short_sha=os.getenv("CI_COMMIT_SHORT_SHA"),
        author=get_git_author(),
    )

    for job in failed_jobs:
        if job["status"] == "failed" and not job["allow_failure"]:
            message += "\n - <{url}|{name}> (stage: {stage}, after {retries} retries)".format(
                url=job["url"], name=job["name"], stage=job["stage"], retries=len(job["retry_summary"]) - 1
            )
        if job["name"].startswith("tests_"):
            for test in get_failed_tests("DataDog/datadog-agent", job):
                message += "\n    - `{}` from package `{}`, owned by *{}*".format(test.name, test.package, test.owners)

    return message


def get_git_author():
    return (
        subprocess.check_output(["git", "show", "-s", "--format='%an'", "HEAD"])
        .decode('utf-8')
        .strip()
        .replace("'", "")
    )


def send_message(recipient, message):
    subprocess.run(["postmessage", recipient, message], check=True)
