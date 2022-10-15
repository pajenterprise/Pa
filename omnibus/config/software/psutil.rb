name "psutil"
default_version "5.7.0"

version "5.7.0" do
  source sha256: "0a53c5260a096aeefdb4060512bbc3cbdec83dd8c20c5114421790b019dc61ff"
end

dependency "python"
dependency "pip"

source url: "https://github.com/giampaolo/psutil/archive/release-#{version}.tar.gz"

relative_path "psutil-release-#{version}"

env = with_embedded_path
env = with_standard_compiler_flags(env)

if linux?
  env = with_glibc_version(env)
  env['CFLAGS'] = "-D_DISABLE_PRLIMIT #{env['CFLAGS']}"
end

build do
  ship_license "https://raw.githubusercontent.com/giampaolo/psutil/master/LICENSE"

  patch source: "psutil-5.6.7-hackadog.patch", env: env

  pip "install --install-option=\"--install-scripts=#{windows_safe_path(install_dir)}/bin\" .", :env => env
end
