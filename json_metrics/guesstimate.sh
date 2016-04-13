echo "Open Javascript console after logging in to getguesstimate.com"
echo "Find your token by running > window.get_profile().token"
echo "GUESS=your token; export GUESS"
curl -vs http://guesstimate.herokuapp.com/spaces -X POST --data @./$1.guess -H "Content-Type: application/json" -H "Authorization: Bearer $GUESS"
