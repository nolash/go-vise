for f in $(ls $1/*.vis); do
	b=$(basename $f)
	b=${b%.*}
	go run ./dev/asm $1/$b.vis > $1/$b.bin
done

for f in $(ls $1/*.txt.orig); do
	b=$(basename $f)
	b=${b%.*}
	#go run ./dev/asm $1/$b.vis > $1/$b.bin
	echo $b
	cp -v $f $1/$b
done
