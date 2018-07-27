#!/usr/bin/env bash

# initial sorting pass for OLD,NEW,and PASSING
cat $1 | grep OLD > got.OLD.out
cat $1 | grep NEW > got.NEW.out
cat $1 | grep PASS > got.PASS.out

anal(){
	while read -r LINE; do

		if [[ "${LINE}" == *"OLD"* ]]; then

			name=$(echo "$LINE" | sed 's/.*name=\([[:graph:]]*\).*/\1/g')
			ruleset=$(echo "$LINE" | sed 's/.*ruleset=\([[:graph:]]*\).*/\1/g')
			if grep -q "$name" "$2"; then
				echo "same: $name/$ruleset"
			else
				echo "uniq: $name/$ruleset"
			fi

		elif [[ "${LINE}" == *"NEW"* ]]; then

			# name=$(echo "$LINE" | sed 's/.*\[NEW\]\([[:graph:]]*\)\:.*/\1/g')
			name=$(echo "$LINE" | sed 's/.*\[NEW\]\([[:graph:]]*\)\/\([[:graph:]]*\)\..*/\2/g')
			# ruleset=$(echo "$LINE" | sed 's/.*ruleset=\([[:graph:]]*\).*/\1/g')
			if grep -q "$name" "$2"; then
				echo "same: $name"
			else
				echo "uniq: $name"
			fi

		else 
			echo "skip: $LINE"
			continue
		fi

	done < "$1"
}

anal got.OLD.out got.NEW.out | sort | uniq | tee got.ON.out
anal got.NEW.out got.OLD.out | sort | uniq | tee got.NO.out

# create grepped and sorted files for convenience
stats(){
	arg=$1
	cat got.$1.out | grep same > got.$1.same.out
	cat got.$1.out | grep uniq > got.$1.uniq.out
}
stats ON
stats NO

# print line counts
printer(){
	echo "---"
	echo "$1 same: $(cat got.$1.same.out | wc -l)"
	echo "$1 uniq: $(cat got.$1.uniq.out | wc -l)"
}
printer ON
printer NO

counts(){
	echo "---"
	echo "$1@Frontier: $(cat got.$1.out | grep Frontier | wc -l)"
	echo "$1@Homestead: $(cat got.$1.out | grep Homestead | wc -l)"
	echo "$1@EIP150: $(cat got.$1.out | grep EIP150 | wc -l)"
	echo "$1@EIP158: $(cat got.$1.out | grep EIP158 | wc -l)"
	echo "$1@Diehard: $(cat got.$1.out | grep Diehard | wc -l)"
	echo "$1@Byzantium: $(cat got.$1.out | grep Byzantium | wc -l)"

}
counts OLD
counts NEW




