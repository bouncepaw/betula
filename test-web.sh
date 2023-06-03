#!/bin/sh

# this script relies on some local files that are not in git, bear with me
make clean > /dev/null
make betula > /dev/null 2> /dev/null
rm testing.betula > /dev/null
killall betula
./betula testing.betula > /dev/null &
sleep 1

LocalInstanceURL='http://localhost:1738'

echo "This is an experimental automated system for testing the WWW interface of Betula."
echo "The instance is located at $LocalInstanceURL. Cross your fingers."
echo

mkdir -p "$HOME/.cache/betula/"
Jar="$HOME/.cache/betula/testing-cookie"
rm "$Jar"

ExpectedContent=''
ExpectedStatus=0
TestName=''
Output=''

Test() {
  TestName=$1
}

ExpectStatus() {
  ExpectedStatus=$1
}

ExpectContent() {
  ExpectedContent=$1
}

ScreamAndShout() {
  echo "Test [$TestName] failed miserably. It is a shame. Please do better."
  echo "Below is the output of cURL."
  echo
  echo "$Output"
  exit 2
}

Check() {
  case "$Output" in
  "HTTP/1.1 $ExpectedStatus"*"$ExpectedContent"*)
    echo "OK [$TestName]"
    ;;
  *)
    ScreamAndShout
  esac
  ExpectedContent=''
}

Call() {
  Addr=$1
  shift
  Output=$(curl -isS "$@" --cookie-jar "$Jar" --cookie "$Jar" "$LocalInstanceURL$Addr")
}

Post() {
  Addr=$1
  shift
  Call "$Addr" -X POST "$@"
}

Get() {
  Addr=$1
  shift
  Call "$Addr" -X GET "$@"
}

Test 'First screen'
ExpectStatus 200
Get '/'
Check

Test 'Register on first screen'
ExpectStatus 303
Post '/register' -F name=bo -F pass=un
Check

Test 'Access non-existent post'
ExpectStatus 404
Get '/1'
Check

Test 'Create a post'
ExpectStatus 303
Post '/save-link' -F url=https://bouncepaw.com -F title=Bouncepaw
Check

Test 'Access the freshly-created post'
ExpectStatus 200
Get '/1'
Check

Test 'Log out'
ExpectStatus 303
Post '/logout'
Check

Test 'Logged out settings'
ExpectStatus 401
Get '/settings'
Check

Test 'Log in'
ExpectStatus 303
Post '/login' -F name=bo -F pass=un
Check

Test 'Save link: both empty'
ExpectStatus 400
ExpectContent 'provide a link'
Post '/save-link'
Check

Test 'Save link: given url empty title'
ExpectStatus 303
Post '/save-link' -F url=https://bouncepaw.com
Check

Test 'Save link: given url to title'
ExpectStatus 303
Post '/save-link' -F title=https://bouncepaw.com
Check

Test 'Save link: both given'
ExpectStatus 303
Post '/save-link' -F url=https://bouncepaw.com -F title=Bouncepaw
Check

Test 'Save link: non-URL title and no URL'
ExpectStatus 400
ExpectContent 'Please, provide a link.'
Post '/save-link' -F url= -F title=Bouncepaw
Check

Test 'Save link: non-URL text to URL'
ExpectStatus 400
ExpectContent 'Invalid link'
Post '/save-link' -F url=Bouncepaw -F title=
Check

Test 'Save link: page with no title, but giving a title'
ExpectStatus 303
Post '/save-link' -F url=https://bouncepaw.com/edge-case/no-title-no-heading -F title='Some title'
Check

Test 'Save link: page with no title, not giving a title'
ExpectStatus 500
ExpectContent 'Title not found'
Post '/save-link' -F url=https://bouncepaw.com/edge-case/no-title-no-heading -F title=
Check

Test 'Save link: headless title'
ExpectStatus 303
Post '/save-link' -F url=https://bouncepaw.com/edge-case/headless-title -F title=
Check
