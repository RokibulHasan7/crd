#!/bin/bash

# change this fields according to yours
packagename=github.com/RokibulHasan7/crd
groupname=rokibul.hasan
versionname=v1alpha1


developmentDir=$(pwd)
k8spath=$HOME/go/src/k8s.io

# getting the code-generator
mkdir -p $k8spath
cd $k8spath && git clone https://github.com/kubernetes/code-generator.git

cd $developmentDir && go get sigs.k8s.io/controller-tools/cmd/controller-gen@v0.11.1
cd $developmentDir && go install sigs.k8s.io/controller-tools/cmd/controller-gen@v0.11.1


# Generate clientset, informers & listers
execDir=$k8spath/code-generator
cd $developmentDir && "${execDir}"/generate-groups.sh all $packagename/pkg/client $packagename/pkg/apis $groupname:$versionname --go-header-file "${execDir}"/hack/boilerplate.go.txt


# Get the dependancies
cd $developmentDir && go mod tidy


# Generate crd manifest file
cd $developmentDir && controller-gen rbac:roleName=controller-perms crd paths=./... output:crd:dir=$developmentDir/manifests output:stdout


