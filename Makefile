release:
	go build -ldflags "-s -w"
	mv DataServer dataserver
doc:
	printf "Assuming godoc is running\n"
	wget -rk --no-parent http://localhost:6060/pkg/github.com/ghsbr/DataServer/
	mv localhost:6060/pkg/github.com/ghsbr/DataServer/ doc
	rm -rf localhost:6060
