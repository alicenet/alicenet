#!/bin/bash
export ANSIBLE_HOST_KEY_CHECKING=FALSE
CLEAN() {
    rm-rf ./$1
}
# Check if ssh key provided
if [ $# -eq 2 ]; then
    echo './create-playbok.sh [SSH_PRIVATE_KEY_FILE] [CHAIN_ID] [TERRAFORM_SA(sa_#)]'
    exit 1
fi

# Check if ssh key exists
if [ ! -f $1 ]; then
    echo 'Could not locate ssh private key file'
    exit 1
fi

CHAIN_ID=$2

if [ $# -lt 3 ]; then
    echo 'Terraform Service Account not specified. (sa_#)'
    exit 1
else
    TERRAFORM_SA=$3
fi

# Add key to access instances && get user
SA_USER=$(gcloud compute os-login ssh-keys add --key-file="$1.pub" | grep "username:" | cut -d':' -f2 | cut -c 2-)
sleep 2s

# Location to write generated playbook to
GEN_DIR="./generated_playbook"

# Remove $GEN_DIR if it exits
if [ -d "$GEN_DIR" ]; then
    CLEAN $GEN_DIR $1
fi

# Create $GEN_DIR
mkdir $GEN_DIR

# Get all tasks from the ansible directory & display to the user
TASKS=($(ls | grep -Ev "reusable_tasks|WIP|*.sh"))
i=0;
echo '[TASKS]'
for T in "${TASKS[@]}"; do
    echo "$i) $T"
    ((i+=1))
done

# Get user selected tasks
echo -n "Select tasks to add: "
read -a SELECTED_TASKS

# if no tasks selected, exit 
if { "${#SELECTED_TASKS[@]}" -eq 0 }; then
    echo 'No tasks selected. Exiting...'
    CLEAN $GEN_DIR $1
    exit 1
fi

echo '- hosts: {{ targets }}"' >> $GEN_DIR/play.yml
for t in "${SELECTED_TASKS[@]}"; do
    if [[ "$t" -ge "${#TASKS[@]}" || "$t" -lt "0" ]]; then
        echo "Invalid task selected. Exiting..."
        CLEAN $GEN_DIR $1
        exit 1
    fi
    echo "- name: Import ${TASKS[$t]}" >> $GEN_DIR/play.yml
    echo " import_playbook: $(pwd)/${TASKS[$t]}/task.yaml" >> $GEN_DIR/play.yaml
done

# Seperator
echo -ne "----------\n"

# Get all targets from gcloud & display to the user
HOSTS=()
INFO=$(gcloud compute instances list --sort-by NAME --format="value(name, zone, networkInterfaces.networkIP)")
while IFS= read -r l; do 
    HOSTS+=($(echo $l | xargs | sed 's/ /:/g'))
done <<< "$INFO"

i=0
echo '[HOSTS]'
for H in "${HOSTS[@]}"; do
    echo "$i) $H"
    ((i+=1))
done

# Get user selected hosts
echo -n "Selected hosts to run tasks: "
read -a SELECTED_HOSTS
# If no hosts selected, exit
if { "${#SELECTED_HOSTS[@]}" -eq 0 }; then
    echo 'No hosts selected. Exiting...'
    CLEAN $GEN_DIR $1
    exit 1
fi

# add each host to the hosts file

echo "[selected_hosts]" >> $GEN_DIR/hosts
HOST_INFO=()
for h in "${SELECTED_HOSTS[@]}"; do
    if [[ "$h" -ge "${#HOSTS[@]}" || "$h" -lt "0" ]]; then
        echo "Invalid host selected. Exiting..."
        CLEAN $GEN_DIR $1
        exit 1
    fi
    echo "${HOSTS[$h]}" |  cut -d":" -f3 >> $GEN_DIR/hosts
    HOST_INFO+=($(echo "${HOSTS[$h]}" | cut -d":" -f 1,3))
done

# Seperator
echo -ne "----------\n"

# Overview confirmation

echo "Playbook Overview"

HEADER=0
for v in "${HOST_INFO[@]}"; do
    if [ $HEADER -eq 0 ]; then
        echo "TARGET @| IP @| TASKS"
        ((HEADER+=1))
        fi
        CONF_NAME=$(echo $v | cut -d':' -f1)
        CONF_IP=$(echo $v | cut -d':' -f2)
        CONF_TASKS=$(echo ${SELECTED_TASKS[@]} | xargs | tr -s ' '', ')
        echo "$CONF_NAME @| $CONF_IP @| $CONF_TASKS"
done | column -s '@' -t

read -n 1 -r -p "Would you like to continue? [y/n] " response
if [[$response =~ ^(n|N|No|no)]]: then
    echo -e"\nAborting...\n"
    CLEAN $GEN_DIR $1
    exit 1
fi

ansible-playbook --key-file $1 --user $SA_USER -i $GEN_DIR/hosts $GEN_DIR/play.yaml -e targets=selected_hosts -e ansible_sa=$SA_USER -e chain_id=$CHAIN_ID

CLEAN $GEN_DIR $1
