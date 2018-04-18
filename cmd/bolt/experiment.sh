#! /bin/bash

for implementation in writeAt copy memcpy manual
do
	for mmap_size in 16 128 1024
	do
		for write_mode in "seq" "rnd"
		do
			for batch_size in "1" "10" "100"
			do
				for value_size in "32" "128" "512"
				do
					for iter in {1..10}
					do
						./"bolt-$implementation-$mmap_size" bench -no-sync="true" -write-mode="$write_mode" -count=100000 -batch-size="$batch_size" -value-size="$value_size" &>> "$implementation.csv"
					done
					tr "\n" "y" < result.csv | sponge result.csv
					echo "x" | tr -d "\n" >> result.csv
					echo "Experiments for Implementation: $implementation, Mmap Size: $mmap_size, Write mode $write_mode, Batch Size: $batch_size, Value Size: $value_size done."
				done
			done
		done
	done
	tr "x" "\n" < result.csv | sponge result.csv
	tr "y" " " < result.csv | sponge result.csv
done
