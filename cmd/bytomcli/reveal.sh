echo "*** create account alice"
alicerootpub=`./bytomcli create-key alice 123456 | sed -n '2p' | cut -d ":" -f 2`
echo "alice root pubkey:" $alicerootpub

alice=`./bytomcli create-account alice $alicerootpub | cut -d ":" -f 2`
echo -e "alice account_id:$alice \n" 

echo "*** create account bob"
bobrootpub=`./bytomcli create-key bob 123456 | sed -n '2p' | cut -d ":" -f 2`
echo "bob root pubkey:" $bobrootpub

bob=`./bytomcli create-account bob $bobrootpub | cut -d ":" -f 2`
echo -e "bob account_id:$bob \n" 

echo "*** create asset gold"
goldrootpub=`./bytomcli create-key glod 123456 | sed -n '2p' | cut -d ":" -f 2`
echo "gold root pubkey:" $goldrootpub

gold=`./bytomcli create-asset glod  $goldrootpub | cut -d ":" -f 2`
echo -e "gold asset_id:$gold \n" 

sleep 20

echo "*** issue asset 10000 gold to alice"
viewcmd="./bytomcli sub-create-issue-tx $alice $gold 10000 123456"
echo $viewcmd
issuecmd="$viewcmd | sed -n '4p' | cut -d "[" -f 3 | cut -d "]" -f 1"
result=`eval $issuecmd`
echo "$result"

#analysis the result
issue=`echo $result | grep code | wc -L`

if [ $issue -ne 0 ]; then
	for i in {1..3}; do
		echo "the $i round Try again!!!"
		t=`eval $issuecmd`
		echo "$t"
		tmpissue=`echo $t | grep code | wc -L`
		if [ $tmpissue -ne 0 ] && [ $i -lt 3 ]; then 
			val=`expr 8 + $i`
			sleep $val
			continue
		elif [ $tmpissue -ne 0 ] && [ $i -eq 3 ]; then
			echo -e "Run command failure!!!\n"
			exit 0
		fi
		echo -e "Run success!!!\n"
		break
	done
else
	echo -e "Run success!!!\n"
fi 

echo "*** spend asset 2000 gold from alice to bob"
viewcmd="./bytomcli sub-spend-account-tx $alice $bob $gold 2000 123456"
echo $viewcmd
spendcmd="$viewcmd | sed -n '5p' | cut -d "[" -f 3 | cut -d "]" -f 1"
result=`eval $spendcmd`
echo "$result"

#analysis the result
spend=`echo $result | grep code | wc -L`

if [ $spend -ne 0 ]; then
	for i in {1..3}; do
		echo "the $i round Try again!!!"
		t=`eval $spendcmd`
		echo "$t"
		tmpspend=`echo $t | grep code | wc -L`
		if [ $tmpspend -ne 0 ] && [ $i -lt 3 ]; then 
			val=`expr 8 + $i`
			sleep $val
			continue
		elif [ $tmpspend -ne 0 ] && [ $i -eq 3 ]; then
			echo -e "Run command failure!!!\n"
			exit 0
		fi
		echo -e "Run success!!!\n"
		break
	done
else
	echo -e "Run success!!!\n"
fi 

#run the command of tools
tools="$GOPATH/src/github.com/bytom/cmd/tools/tools"

#genenate the sha3 hash for string
echo "*** genenate the sha3 hash for string"
sha3hash="$GOPATH/src/github.com/bytom/exp/ivy/tools/tools"
str="abcd"
viewcmd="$sha3hash $str"
echo $viewcmd
hash=`eval $viewcmd | sed -n '2p' | cut -d ":" -f 2`
echo -e "the hash of pubkey:$hash\n" 

#genenate the bytecode for the contract of RevealPreimage
echo "*** genenate the bytecode for the contract of RevealPreimage"
ivy="$GOPATH/src/github.com/bytom/exp/ivy/ivy"
viewcmd="$ivy RevealPreimage $hash"
echo $viewcmd
bytecode=`eval $viewcmd | sed -n '2p'`
echo -e "contract bytecode:$bytecode\n"

echo "*** create contract to alice"
viewcmd="$tools contract $alice $bytecode"
echo $viewcmd
contract=`eval $viewcmd | sed -n '2p'`
echo -e "contract response:\n$contract\n"

#exec the lock operation
echo "*** exec the lock operation for alice"
viewcmd="$tools lock $alice $gold 123456 88 $bytecode"
echo $viewcmd
lockcmd="$viewcmd | sed -n '6p' | cut -d "[" -f 3 | cut -d "]" -f 1"
result=`eval $lockcmd`
echo "$result"

#analysis the result
issue=`echo $result | grep code | wc -L`

if [ $issue -ne 0 ]; then
	for i in {1..3}; do
		echo "the $i round Try again!!!"
		t=`eval $lockcmd`
		echo "$t"
		tmplock=`echo $t | grep code | wc -L`
		if [ $tmplock -ne 0 ] && [ $i -lt 3 ]; then 
			val=`expr 10 + $i`
			sleep $val
			continue
		elif [ $tmplock -ne 0 ] && [ $i -eq 3 ]; then
			echo -e "Run command failure!!!\n"
			exit 0
		fi
		echo -e "Run success!!!\n"
		break
	done
else
	echo -e "Run success!!!\n"
fi 

echo "*** exec the unlock operation for alice"
read -p "ouputid:" ouputid
echo $ouputid
viewcmd="$tools unlockRevPreimage $ouputid $alice $gold 123456 88 abc $str"
echo $viewcmd
result=`eval $viewcmd`
echo "$result"

