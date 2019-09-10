#!/bin/bash

# Licensed Materials - Property of IBM
# 5737-E67
# (C) Copyright IBM Corporation 2016, 2019 All Rights Reserved
# US Government Users Restricted Rights - Use, duplication or disclosure restricted by GSA ADP Schedule Contract with IBM Corp.

YEAR=2019

#LINE1="${COMMENT_PREFIX}Licensed Materials - Property of IBM"
CHECK1=" Licensed Materials - Property of IBM"
#LINE2="${COMMENT_PREFIX}(c) Copyright IBM Corporation ${YEAR}. All Rights Reserved."
CHECK2=" 5737-E67"
#LINE3="${COMMENT_PREFIX}Note to U.S. Government Users Restricted Rights:"
CHECK3=" (C) Copyright IBM Corporation 2016, ${YEAR} All Rights Reserved"
#LINE4="${COMMENT_PREFIX}Use, duplication or disclosure restricted by GSA ADP Schedule"
CHECK4=" US Government Users Restricted Rights - Use, duplication or disclosure restricted by GSA ADP Schedule Contract with IBM Corp."

#LIC_ARY to scan for
LIC_ARY=("$CHECK1" "$CHECK2" "$CHECK3" "$CHECK4")
LIC_ARY_SIZE=${#LIC_ARY[@]}

#Used to signal an exit
ERROR=0
TOTAL_ERROR=0

echo "##### Copyright check #####"
#Loop through all files. Ignore .FILENAME types
for f in `find . -type f ! -iname ".*" ! -path "./pkg/client/*" ! -path "./vendor/*" ! -iname "*_generated*" ! -path "./build-harness/*"`; do
  if [ ! -f "$f" ] || [ "$f" = "./build-tools/copyright-check.sh" ]; then
    continue
  fi

  FILETYPE=$(basename ${f##*.})
  case "${FILETYPE}" in
  	js | sh | java | rb)
  		COMMENT_PREFIX=""
  		;;
  	*)
      continue
  esac

  #Read the first 10 lines, most Copyright headers use the first 6 lines.
  HEADER=`head -10 $f`
  printf " Scanning $f . . . "

  #Check for all copyright lines
  for i in `seq 0 $((${LIC_ARY_SIZE}+1))`; do
    #Add a status message of OK, if all copyright lines are found
    if [ $i -eq ${LIC_ARY_SIZE} ]; then
      printf "OK\n"
    else
      #Validate the copyright line being checked is present
      if [[ "$HEADER" != *"${LIC_ARY[$i]}"* ]]; then
        TOTAL_ERROR=$((TOTAL_ERROR+1))
        printf "Missing copyright\n  >>Could not find [${LIC_ARY[$i]}] in the file $f\n"
        printf "Appending copyright to $f\n"
        
        file=${PWD}$(echo $f | cut -c2-)

        if [[ ${file:(-2)} = "sh" ]]
        then 
            TOTAL_ERROR=$((TOTAL_ERROR-1))
            ./build/update-copyright.sh $file "#"
            break
        # else 
        #     if [[ ${file:(-2)} = "go" ]]
        #     then 
        #       TOTAL_ERROR=$((TOTAL_ERROR-1))
        #       ./build/update-copyright.sh $file "//"
        #       break
        #     fi uncomment and add to line 36 for go goodness :))))))
        fi
      fi
    fi
    done
done

if [[ $TOTAL_ERR -gt 0 ]]
then 
    ERROR=1
    break

fi


echo "##### Copyright check ##### ReturnCode: ${ERROR}"
exit $ERROR
