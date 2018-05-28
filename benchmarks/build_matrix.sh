#!/bin/bash -ex

mkdir -p bin

IMAGES=${IMAGES-golang:1.4 golang:1.5 golang:1.6 golang:1.7 golang:1.8 golang:1.9 golang:1.10 gcc:5 gcc:6 gcc:7 gcc:8}
BENCHMARKS=${BENCHMARKS-$(find * -maxdepth 0 -type d)}
SYSTEMS=${SYSTEMS-linux darwin windows}

for IMAGE in $IMAGES; do
  for BENCHMARK in $BENCHMARKS; do
    if [[ $BENCHMARK == "bin" ]]; then continue; fi
    if [[ $BENCHMARK == "cpu_bound" ]]; then continue; fi
    if [[ $BENCHMARK == "example" ]]; then continue; fi
    for GOOS in $SYSTEMS; do
      if [[ $IMAGE == gcc* && $GOOS != "linux" ]]; then
	continue
      fi

      if [[ $IMAGE == golang:1.4 && $GOOS != "linux" ]]; then
	PREAMBLE='cd /usr/src/go/src && ./make.bash &&'
      else
	PREAMBLE=
      fi

      if [[ $GOOS == "windows" ]]; then
	EXT=".exe"
      else
	EXT=""
      fi

      docker run --rm -v "$PWD/$BENCHMARK:/app" -e "GOOS=$GOOS" $IMAGE bash -c "$PREAMBLE cd /app && go build && chown 1000:1000 app$EXT"
      mv $BENCHMARK/app$EXT bin/$BENCHMARK.$IMAGE.$GOOS$EXT
    done
  done
done
