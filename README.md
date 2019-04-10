# First playground with MongoDB in GO programming

## Quick Start

### create project

``` bash
mkdir -p $GOPATH/src/github/{your username}/{project name}
```

### Write main.go

``` bash
# build the project
go build
# run the execution file
./project_name
```

### Install mongodb driver

``` bash
go get go.mongodb.org/mongo-driver/mongo
```

### Install other dependecies, dotenv for .env file, mux for router

``` bash
go get github.com/joho/godotenv
go get github.com/
go get -u github.com/rs/zerolog/log
```

## Version

1.0.0

## Contributing

1. Fork it
2. Create your feature branch (`git checkout -b my-new-feature`)
3. Commit your changes (`git commit -am 'Add some feature'`)
4. Push to the branch (`git push origin my-new-feature`)
5. Create new Pull Request
