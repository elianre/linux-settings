#/usr/bin/env sh
sudo apt-get install vim vim-gnome ctags cmake perl curl build-essential git libgtk2.0-dev pkg-config python-dev python-numpy libavcodec-dev libavformat-dev libswscale-dev autoconf automake libtool
sudo apt-get install wmctrl xdotool kdiff3

wget http://python.org/ftp/python/3.3.0/Python-3.3.0.tgz
tar -xzf Python-3.3.0.tgz
cd Python-3.3.0
./configure --prefix=/opt/python3.3
make  
sudo make install

curl -L http://install.perlbrew.pl | bash
perlbrew install perl-5.8.1

sudo add-apt-repository ppa:shutter/ppa
sudo apt-get update
sudo apt-get install shutter


sudo cp vimrc.local /etc/vim/vimrc.local

git config --global user.email "renliang87@gmail.com"
git config --global user.name "elianre"
