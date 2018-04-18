#! /bin/bash

for implementation in -no-mmap-write copy
do
		for batch_size in "1" "10" "100"
		do
			for value_size in "32" "128" "512"
			do
				for iter in {1..10}
				do
					./bolt bench -no-sync="true" -go-file-mode -count=100000 -batch-size="$batch_size" -value-size="$value_size" &>> "$implementation.csv"
				done
				tr "\n" "y" < result.csv | sponge result.csv
				echo "x" | tr -d "\n" >> result.csv
				echo "Experiments for Implementation: $implementation, Mmap Size: $mmap_size, Write mode $write_mode, Batch Size: $batch_size, Value Size: $value_size done."
			done
		done
	tr "x" "\n" < result.csv | sponge result.csv
	tr "y" " " < result.csv | sponge result.csv
done
