name "aerospike"
default_version "3.10.0"

dependency "pip"

build do
  # The binary wheels on PyPI are not yet compatible with OpenSSL 1.1.0+, see:
  # https://github.com/aerospike/aerospike-client-python/issues/214#issuecomment-385451007
  # https://github.com/aerospike/aerospike-client-python/issues/227#issuecomment-423220411
  command "git clone https://github.com/aerospike/aerospike-client-c.git #{install_dir}/embedded/lib/aerospike"

  # https://github.com/aerospike/aerospike-client-python/blob/master/BUILD.md#building-on-an-unsupported-linux-distro
  command "git clone https://github.com/aerospike/aerospike-lua-core.git #{install_dir}/embedded/lib/aerospike/aerospike-lua-core"

  # This needs to be kept in sync with whatever the Python library was built with.
  # For example, version 3.10.0 was built with version 4.6.10 of the C library, see:
  # https://github.com/aerospike/aerospike-client-python/blob/3.10.0/setup.py#L32-L33
  command "cd #{install_dir}/embedded/lib/aerospike && git checkout 4.6.10"

  command "cd #{install_dir}/embedded/lib/aerospike && git submodule update --init"

  env = {
    "LDFLAGS" => "-L#{install_dir}/embedded/lib -I#{install_dir}/embedded/include",
    "CFLAGS" => "-L#{install_dir}/embedded/lib -I#{install_dir}/embedded/include",
    "LD_RUN_PATH" => "#{install_dir}/embedded/lib",
    "DOWNLOAD_C_CLIENT" => "0",
    "AEROSPIKE_C_HOME" => "#{install_dir}/embedded/lib/aerospike",
    "AEROSPIKE_LUA_PATH" => "#{install_dir}/embedded/lib/aerospike/aerospike-lua-core/src",
  }

  command "cd #{install_dir}/embedded/lib/aerospike && make clean", :env => env
  command "cd #{install_dir}/embedded/lib/aerospike && make", :env => env

  pip "install --no-binary aerospike aerospike==#{version}", :env => env
end
