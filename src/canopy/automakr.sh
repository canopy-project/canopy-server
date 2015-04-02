while :
do
    OUTPUT=$(make 2>&1)
    clear
    echo "$OUTPUT"
    sleep 1
done
