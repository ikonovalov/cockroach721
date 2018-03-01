DC_FILE=docker/docker-compose.yml
ETH_TOOLS=${GOPATH}/src/github.com/ethereum/go-ethereum/build/bin
PACKAGE=cockroach

solc-upgrade:
	docker pull ethereum/solc:stable
solc-compile-shutdown:
	 docker-compose -f ${DC_FILE} down
run: build
	./bin/dapp
build:
	go build -o bin/dapp .

solc-compile:
	docker-compose -f ${DC_FILE} up && ls -l contracts/gen && docker-compose -f ${DC_FILE} down
gen-bind:
	${ETH_TOOLS}/abigen -abi ./contracts/gen/NonFungibleToken.abi -bin ./contracts/gen/NonFungibleToken.bin -pkg ${PACKAGE} -lang go -out contracts/bind/cockroach721.go -type CockroachToken
check-security:
	myth -c `cat ./contracts/gen/NonFungibleToken.bin` -x