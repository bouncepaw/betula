#!/bin/sh

LocalInstanceURL='http://localhost:1738'

echo "This is an experimental automated system for testing the WWW interface of Betula."
echo "The instance is located at $LocalInstanceURL. Cross your fingers."

ExpectedStatus=0
ActualStatus=0
TestName=''
Output=''

Test() {
  TestName=$1
}

ExpectStatus() {
  ExpectedStatus=$1
}

Check() {
  if [ "$ExpectedStatus" = "$ActualStatus" ]
  then
    echo "Test [$TestName] successful!"
  else
    echo "Test [$TestName] failed miserably. It is a shame. Please do better."
    echo "Below is the output of cURL."
    echo
    echo "$Output"
    exit 2
  fi
}

Call() {
  Addr=$1
  shift
  Output=$(curl -sS "$@" "$LocalInstanceURL""$Addr")
}

Post() {
  Addr=$1
  shift
  Call "$Addr" -X POST "$@"
}


Test 'Log in'
ExpectStatus 200
  # Such is my test password.
Post '/login' -F name=bo -F password=un
Check

