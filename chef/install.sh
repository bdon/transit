# check for chef-solo binary
# right now using ruby-1.9.3 from repo, which is not latest patch

chef_binary=/var/lib/gems/1.9.1/gems/chef-11.4.0/bin/chef-solo

if ! test -f "$chef_binary"; then
    echo "Bootstrapping vanilla system..."
    export DEBIAN_FRONTEND=noninteractive
    aptitude update && aptitude install -y ruby1.9.3 ruby1.9.1-dev make &&
    sudo gem install --no-rdoc --no-ri chef --version 11.4 &&
    sudo gem install --no-rdoc --no-ri bundler
fi &&

"$chef_binary" -c solo.rb -j solo.json

