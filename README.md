# hydra-sample

## Prerequisites

__Ubuntu, ubuntu-make__

Scripts below are valid for Ubuntu, you may need to modify them for other OSes.

See https://wiki.ubuntu.com/ubuntu-make for details on `umake` usage.

To install `umake` on Ubuntu just use

```
sudo apt-get install ubuntu-make
```

__Go (Golang)__

Use `umake` to install or update Golang

```
umake go
```

__Glide__

We use Glide to dependency management. You may install it as below, or just use its
configuration file `./backend/glide.yaml` as a reference of required dependencies.

```
sudo add-apt-repository ppa:masterminds/glide && sudo apt-get update
sudo apt-get install glide
```

__NPM, nodejs__

Use `umake` to install or update nodejs

```
umake nodejs
```


## Build project from sources

```
WRK_DIR=$GOPATH/src/github.com/letsrock-today
mkdir -p $WRK_DIR
cd $WRK_DIR
git clone https://github.com/letsrock-today/hydra-sample.git
cd $WRK_DIR/hydra-sample/backend
glide up

cd $WRK_DIR/hydra-sample/ui-web
npm install
npm run dist

cd $WRK_DIR/hydra-sample/backend
go run main.go

```


