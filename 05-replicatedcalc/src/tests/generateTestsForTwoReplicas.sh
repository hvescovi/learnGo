#!/bin/bash
#generator of test cases

if [ -z $1 ]; then
  echo "informs number of requests"
  echo "usage: generateTests <numberOfRequests> <numberOfClients> <client-ID-start-number> <operator>"
  echo "if operator is not informed, they will be generated in a round-robin style"
  exit 0
fi

if [ -z $2 ]; then
  echo "informs number of clients"
  exit 0
fi

if [ -z $3 ]; then
  echo "informs the ID/number of the first client"
  exit 0
fi

# use "let" to get parameters as numbers
let nRequests=$1
let nClients=$2
let firstClient=$3

#echo " generating testfile "

# temporary file
if [ -f testfile.sh ]
then
  rm testfile.sh
fi
touch testfile.sh

nop=1
cmd="none"

let cIni=1+firstClient-1
let cFim=nClients+firstClient-1
#echo $cIni
#echo $cFim

let port=8001

for ((r=1; r<=$nRequests; r++))
do
  for ((c=$cIni; c<=$cFim; c++))
  do
    id1="c$c"
    id2="r$r"
    id=$id1$id2

    if [ ! -z $4 ]
    then
      cmd=$4
    else
      if (( nop == 1 ));
      then
        cmd="get"
      fi
      if (( nop == 2 ));
      then
        cmd="inc"
      fi
      if (( nop == 3 ));
      then
        cmd="dou"
      fi

      nop=$nop+1
      if (( nop == 4 ));
      then
        nop=1
      fi

      let port=$port+1
      if (( port == 8003 ));
      then
        port=8001
      fi
      sp="$port"
    fi

#    echo "curl http://192.168.15.3:32550/$id/$cmd" #>> testfile.sh
    echo "curl localhost:$sp/$id/$cmd/3" #>> testfile.sh
  done
done

#mix the file
#java -jar textRaffle.jar testfile.sh > testmix.sh

#insert the header
#echo "echo clientId size operation method t0 t1 t2 t3 t4 t5 t6 t7 t8 end delay" > head.txt
#cat testmix.sh >> head.txt

# final adjustments
#rm testmix.sh
#mv head.txt test.sh
#chmod u+x test.sh

#chmod u+x testmix.sh

#echo "done"
