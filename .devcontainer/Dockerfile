# [Choice] Go version (use -bullseye variants on local arm64/Apple Silicon): 1, 1.19, 1.18, 1-bullseye, 1.19-bullseye, 1.18-bullseye, 1-buster, 1.19-buster, 1.18-buster
# ARG VARIANT=1.21-bullseye
# FROM mcr.microsoft.com/devcontainers/go:1-${VARIANT}
FROM mcr.microsoft.com/devcontainers/base:debian-13

# [Optional] Uncomment this line to install global node packages.
# RUN su vscode -c "source /usr/local/share/nvm/nvm.sh && npm install -g <your-package-here>" 2>&1

RUN apt update && DEBIAN_FRONTEND=nointeractive apt install -y --no-install-recommends locales tzdata lv zsh vim
RUN echo "ja_JP.UTF-8 UTF-8" >> /etc/locale.gen
RUN locale-gen
RUN ln -sf /usr/share/zoneinfo/Asia/Tokyo /etc/localtime
RUN echo "Asia/Tokyo" > /etc/timezone

RUN chsh -s /usr/bin/zsh vscode

USER vscode
RUN curl https://mise.run | bash
RUN echo 'eval "$(/home/vscode/.local/bin/mise activate zsh)"' >> /home/vscode/.zshrc