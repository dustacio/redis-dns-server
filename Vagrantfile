# -*- mode: ruby -*-
# vi: set ft=ruby :

VAGRANTFILE_API_VERSION = "2"

Vagrant.configure(VAGRANTFILE_API_VERSION) do |config|
  config.vm.synced_folder ".", "/vagrant"
  config.vm.box_url = "https://cloud-images.ubuntu.com/vagrant/trusty/current/trusty-server-cloudimg-amd64-vagrant-disk1.box"
  config.vm.box = "ubuntu-14.04-amd64"
  config.vm.network "private_network", type: :dhcp, auto_config: true
  config.vm.provision "shell" do |s|
    s.inline = "cd /vagrant && bash scripts/vagrant-up.sh"
  end

  config.vm.provider "virtualbox" do |v|
    v.memory = 1024
    v.cpus = 2
    v.gui = false
  end
end
