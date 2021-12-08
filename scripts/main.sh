#!/bin/bash
re='^[0-9]+$'

BUILD() {
    make build
}

PRE_CHECK () {
    # Check if madnet binary exists
    if [ ! -f "./madnet" ]; then
        BUILD
    fi
    # Check that the generated directory exists
    if [[ "$1" != "init" ]] && [ ! -d "./scripts/generated" ]; then
        echo -e "Need to initialize before continuing";
        exit 1;
    fi
    # Check all required non builtins exist
    COMMANDS=("ethkey" "jq" "hexdump")
    for c in ${COMMANDS[@]}; do
        if ! command -v $c &> /dev/null;
        then
            echo -e "$c is required, but not installed"
            exit 1
        fi
    done
}

CLEAN_UP () {
    # Reset Folder
    rm -rf ./scripts/generated
    if [[ "$1" == "all" ]]; then
        exit 0
    fi
    # Init
    mkdir ./scripts/generated
    mkdir ./scripts/generated/stateDBs
    mkdir ./scripts/generated/config
    mkdir ./scripts/generated/keystores
    mkdir ./scripts/generated/keystores/keys
    touch ./scripts/generated/keystores/passcodes.txt
    cp ./scripts/base-files/genesis.json ./scripts/generated/genesis.json
    cp ./scripts/base-files/0x546f99f244b7b58b855330ae0e2bc1b30b41302f ./scripts/generated/keystores/keys
    echo -e "0x546F99F244b7B58B855330AE0E2BC1b30b41302F=abc123" > ./scripts/generated/keystores/passcodes.txt
}

CREATE_CONFIGS () {
    # Vars
    LA=4242
    PA=4343
    DA=4444
    LSA=8884
    # Check that number of validators is valid
    if ! [[ $1 =~ $re ]] || [[ $1 -lt 4 ]] || [[ $1 -gt 32 ]]; then
        echo -e "Invalid number of validators [4-32]"
        exit 1
    fi
    if [ -f "./scripts/generated/genesis.json" ]; then
        echo -e "Generated files already exist, run clean"
        exit 1
    fi
    CLEAN_UP
    # Loop through and create all essentail validator files
    for (( l = 1; l <= $1; l++ )) ; do
        ADDRESS=$(ethkey generate --passwordfile ./scripts/base-files/passwordFile | cut -d' ' -f2)
        PK=$(hexdump -n 16 -e '4/4 "%08X" 1 "\n"' /dev/urandom)
        sed -e 's/defaultAccount = .*/defaultAccount = \"'"$ADDRESS"'\"/' ./scripts/base-files/baseConfig |
        sed -e 's/rewardAccount = .*/rewardAccount = \"'"$ADDRESS"'\"/' |
        sed -e 's/listeningAddress = .*/listeningAddress = \"0.0.0.0:'"$LA"'\"/' |
        sed -e 's/p2pListeningAddress = .*/p2pListeningAddress = \"0.0.0.0:'"$PA"'\"/' |
        sed -e 's/discoveryListeningAddress = .*/discoveryListeningAddress = \"0.0.0.0:'"$DA"'\"/' |
        sed -e 's/localStateListeningAddress = .*/localStateListeningAddress = \"0.0.0.0:'"$LSA"'\"/' |
        sed -e 's/passcodes = .*/passcodes = \"scripts\/generated\/keystores\/passcodes.txt\"/' |
        sed -e 's/keystore = .*/keystore = \"scripts\/generated\/keystores\/keys\"/' |
        sed -e 's/stateDB = .*/stateDB = \"scripts\/generated\/stateDBs\/validator'"$l"'\/\"/' |
        sed -e 's/privateKey = .*/privateKey = \"'"$PK"'\"/' > ./scripts/generated/config/validator$l.toml
        echo "$ADDRESS=abc123" >> ./scripts/generated/keystores/passcodes.txt
        mv ./keyfile.json ./scripts/generated/keystores/keys/$ADDRESS
        jq '.alloc += {"'"$(echo $ADDRESS | cut -c3-)"'": {balance:"80000000000000000"}}' ./scripts/generated/genesis.json > ./scripts/generated/genesis.json.tmp && mv ./scripts/generated/genesis.json.tmp ./scripts/generated/genesis.json
        ((LA=LA+1))
        ((PA=PA+1))
        ((DA=DA+1))
        ((LSA=LSA+1))
    done
}

LIST () {
    # List each of the validators
    COUNTER=1
    for f in $(ls ./scripts/generated/config | xargs); do
        echo -e "$COUNTER : $f"
        COUNTER=$((COUNTER+1))
    done
}

CHECK_EXISTING() {
    # Check if validator exists
    if ! [[ $1 =~ $re ]] || [[ $1 -le 0 ]]; then
        echo -e "Invalid number"
        exit 1
    fi
    if [ ! -f "./scripts/generated/config/validator$1"]; then
        echo -e "Validator $1 does not exist"
        exit 1
    fi
}

RUN_VALIDATOR() {
    # Run a validator
    CHECK_EXISTING $1
    ./madnet --config ./scripts/generated/config/validator$1.toml validator
}

STATUS() {
    # Check validator status
    CHECK_EXISTING $1
    ./madnet --config ./assets/config/validator$1.toml utils
}

# init # - initalize validators directory files
# geth - start geth
# bootnode - start bootnode
# deploy - deploy necessary contracts
# validator # - run a validator by number
# ethdkg - launch ethdkg
# deposit - run a deposit to the owner toml
# unregister - unregister all the validators
# list - list the validators
# status # - get the status of a validator
# clean - remove all generated files

PRE_CHECK $1
case $1 in
    init)
        CREATE_CONFIGS $2
    ;;
    geth)
        ./scripts/base-scripts/geth-local.sh
    ;;
    bootnode)
        ./scripts/base-scripts/bootnode.sh
    ;;
    deploy)
        ./scripts/base-scripts/deploy.sh && ./scripts/base-scripts/approvetokens.sh && ./scripts/base-scripts/transfertokens.sh && ./scripts/base-scripts/register.sh
    ;;
    validator)
        RUN_VALIDATOR $2
    ;;
    ethdkg)
        ./scripts/base-scripts/ethdkg.sh
    ;;
    deposit)
        ./scripts/base-scripts/deposit.sh
    ;;
    unregister)
        ./scripts/base-scripts/unregister.sh
    ;;
    list)
        LIST
    ;;
    status)
        STATUS $2
    ;;
    clean)
        CLEAN_UP "all"
    ;;
    *)
        echo -e "Unknown argument!"
        echo -e "init # | geth | bootnode | deploy | validator # | ethdkg | deposit | unregister | list | status | clean"
        exit 1;
esac
exit 0
