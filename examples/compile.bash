for f in $(ls $1); do
	b=$(basename $f)
	b=${b%.*}
	go run ./dev/asm $1/$b.vis > $1/$b.bin
done
