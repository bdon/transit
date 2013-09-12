package 'curl'
package 'build-essential'
package 'libssl-dev'
package 'libsqlite3-0'
package 'libsqlite3-dev'
package 'python-software-properties'

directory "/var/www/tilestream" do
  recursive true
  owner "bdon"
  group "bdon"
end

directory "/usr/share/tilestream" do
  recursive true
  owner "bdon"
  group "bdon"
end

execute "install node" do
  command "apt-add-repository ppa:chris-lea/node.js &&
           apt-get update &&
           apt-get install -y nodejs"
  not_if { File.exists?("/usr/bin/node") }
end

execute "install tilestream" do
  command "cd /var/www/tilestream &&
           git clone https://github.com/mapbox/tilestream.git &&
           cd tilestream &&
           npm install"
  not_if { File.exists?("/var/www/tilestream/tilestream") }
end

cookbook_file "/etc/init/tilestream.conf" do
  source "tilestream.conf"
end
