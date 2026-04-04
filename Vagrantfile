Vagrant.configure("2") do |config|
  config.vm.box = ENV.fetch("VAGRANT_BOX", "bento/ubuntu-22.04")
  config.vm.hostname = "bypass-lab"

  config.vm.network "forwarded_port", guest: 80, host: 8080, auto_correct: true
  config.vm.network "public_network",
    bridge: ENV.fetch("VAGRANT_BRIDGE", "Ethernet"),
    use_dhcp_assigned_default_route: true

  config.vm.provider "virtualbox" do |vb|
    vb.name = "bypass-403-go-lab"
    vb.memory = 2048
    vb.cpus = 2
  end

  config.vm.provider "vmware_desktop" do |vmware|
    vmware.vmx["memsize"] = "2048"
    vmware.vmx["numvcpus"] = "2"
  end

  config.vm.provision "shell", path: "vagrant/provision.sh"
end
