find . -type f -name '*.sh' | while read FILE ; do
    newfile="$(echo ${FILE} |sed -e 's/sql/plasma/')" ;
    mv "${FILE}" "${newfile}" ;
done
