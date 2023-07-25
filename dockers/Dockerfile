FROM archlinux:latest

WORKDIR /root/

RUN pacman -Syyu --noconfirm 

# install basic libs and tools
RUN pacman -Sy --noconfirm base-devel nasm bc

# install libs and tools are necessary for Linux
RUN pacman -Sy --noconfirm flex bison libelf-dev libssl-dev kmod

# install build tools
RUN pacman -Sy --noconfirm ninja-build cmake

# install network tools and "downloaders"
RUN pacman -Sy --noconfirm netcat git curl net-tools wget

# install debug tools
RUN pacman -Sy --noconfirm gdb

# install qemu
RUN pacman -Sy --noconfirm qemu kvm

# install required tools for creating images
RUN pacman -Sy --noconfirm sudo debootstrap

# install python ruby environment
RUN pacman -Sy --noconfirm python3-dev python3-pip ipython3 ruby-dev

# install vim with .vimrc
RUN pacman -Sy --noconfirm vim &&\
    git clone --depth=1 https://github.com/amix/vimrc.git ~/.vim_runtime &&\
    bash ~/.vim_runtime/install_awesome_vimrc.sh

# install zshrc
RUN pacman -Sy --noconfirm zsh && git clone --depth=1 https://github.com/robbyrussell/oh-my-zsh.git /tmp/oh-my-zsh && cat /tmp/oh-my-zsh/tools/install.sh | zsh &&\
    sed -i "1,$ s/ZSH_THEME=\"robbyrussell\"/ZSH_THEME=\"ys\"/g" ~/.zshrc && \
    chsh -s /bin/zsh

# install tmux with .tmux
RUN pacman -Sy --noconfirm tmux && \
    git clone https://github.com/gpakosz/.tmux.git && \
    ln -s -f .tmux/.tmux.conf && cp .tmux/.tmux.conf.local .


CMD /bin/zsh

