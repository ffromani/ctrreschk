all: binaries

binaries: ctrreschk

outdir:
	@mkdir -p _out

ctrreschk: outdir
	go build -v -o _out/ctrresck cmd/ctrreschk/main.go

