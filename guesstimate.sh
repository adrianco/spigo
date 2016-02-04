echo "Set GUESS to window.my_api_token from Javascript console after logging in to getguesstimate.com"
curl -vs http://guesstimate.herokuapp.com/spaces -X POST --data @./$1.guess -H "Content-Type: application/json" -H "Authorization: Bearer $GUESS"
