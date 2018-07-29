#!/usr/bin/env bash

gethout="$1"
parityout="$2"

while read -r LINE; do
	if [[ "${LINE}" == *"0x"* ]]; then
		# remember, there will be want+got outs (hopefully), so the match should match 2x
		# capture first match
		match1=$(echo "$LINE" | sed 's/.*0x\([[:graph:]]*\).*/\1/')
		# capture second match
		match2=$(echo "$LINE" | sed 's/.*0x\([[:graph:]]*\).*/\1/')


		# look for first and second matches in parity test.log file
		if [[ "${#match1}" -gt 19 ]] && grep -q "$match1" "$parityout"; then
			echo "matched (geth): $LINE"
			continue
		elif [[ "${#match2}" -gt 19 ]] &&  grep -q "$match2" "$parityout"; then
			echo "matched (geth): $LINE"
		fi
	fi
done < "$gethout"
