#!/bin/sh
curl https://raw.githubusercontent.com/RedHat-EMEA-SSA-Team/workshop-operator/2.12/htpasswd/htpasswd > htpasswd
oc create secret generic htpasswd --from-file=htpasswd=htpasswd -n openshift-config
oc apply -f https://raw.githubusercontent.com/RedHat-EMEA-SSA-Team/workshop-operator/2.12/htpasswd/oauth_htpasswd_provider -n openshift-config
oc adm policy add-cluster-role-to-user cluster-admin opentlc-mgr
rm users.htpasswd
