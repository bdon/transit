package 'nginx'
package 'htop'
package 'g++'

execute "install GEOS 3.3.8" do
   command "mkdir /tmp/geos && \
cd /tmp/geos &&
wget http://download.osgeo.org/geos/geos-3.3.8.tar.bz2 &&
tar xvfj geos-3.3.8.tar.bz2 &&
cd geos-3.3.8 &&
./configure &&
make &&
make install
"
  not_if { File.exists?("/usr/local/include/geos_c.h") }
end

execute "install go" do
  command "mkdir /tmp/go &&
cd /tmp/go &&
wget https://go.googlecode.com/files/go1.1.1.linux-amd64.tar.gz &&
tar -C /usr/local -xzf go1.1.1.linux-amd64.tar.gz"
  not_if { File.exists?("/usr/local/go") }
end

cookbook_file "/etc/profile.d/go.sh" do
  mode 0744
  source "go.sh"
end

cookbook_file "/etc/init/api.conf" do
  source "api.conf"
end

directory "/usr/local/serve" do

end

cookbook_file '/etc/nginx/nginx.conf' do
  source 'nginx.conf'
  mode 0644
  owner 'root'
  group 'root'
  notifies :reload, "service[nginx]"
end

cookbook_file '/etc/motd.tail' do
  source 'motd.tail'
  mode 0644
  owner 'root'
  group 'root'
end

service 'nginx' do
  supports :status => true, :restart => true, :reload => true
  action :start
end

directory '/var/www' do
  owner 'bdon'
  group 'bdon'
  mode 00755
  action :create
end

directory '/var/www/bdon.org' do
  owner 'bdon'
  group 'bdon'
  mode 00755
  action :create
end
