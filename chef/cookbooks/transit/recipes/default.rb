package 'nginx'
package 'htop'
package 'libgeos-3.2.2'
package 'libgeos-dev'

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
