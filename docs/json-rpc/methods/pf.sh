#!/usr/bin/env bash

# OIFS=$IFS
# IFS=$'\n'
section=""
demarc="^\*\*\*"
header='####'
while read -r li; do
	echo "LI: $li"
	section+="$li
"


	if [[ $li =~ $demarc ]]; then
		title="$(echo $section | grep $header | head -n1 | cut -d' ' -f2)"
		echo "TITLE: $title"
		mparent="$(echo $title | cut -d'_' -f1)"
		echo "MPARENT: $mparent"
		mmethod="$(echo $title | cut -d'_' -f2)"
		echo "METHOD: $method"

		mkdir -p "$mparent"

		# remove demarc last line
		# section="$(echo $section | sed '$ d')"
	echo "SECTION:
	$section"

		echo "$section" > "$mparent/$mmethod.md"
		section=""
	fi
done < JSON-RPC.md

for f in ./**/*.md; do
	sed -i '$ d' "$f"
done

# IFS=$OIFS


