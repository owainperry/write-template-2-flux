# write-template-2-flux

    REPO_PATH=/tmp/$REPO_NAME
    ROLENAME="arn:aws:iam::$AWS_ACCOUNTNUMBER:role/kustomize-controller-service-account-$ENV_NAME"
    git clone https://${GITHUB_TOKEN}:x-oauth-basic@github.com/$OWNER/$REPO_NAME.git $REPO_PATH
    mkdir -p $REPO_PATH/clusters/$ENV_NAME/flux-system
    FILE=$REPO_PATH/clusters/$ENV_NAME/flux-system/kustomization.yaml
    VAR=$(cat <<EOM
    patches:
    - patch: |
        - op: add
          path: "/metadata/annotations/eks.amazonaws.com~1role-arn"
          value: $ROLENAME
      target:
        kind: ServiceAccount
        name: "kustomize-controller"
    EOM
    )
    echo "$VAR" >> $FILE

    ROLENAME="arn:aws:iam::$AWS_ACCOUNTNUMBER:role/cluster-autoscaler-service-account-$ENV_NAME"
    FILE=$REPO_PATH/clusters/$ENV_NAME/kustomization-cluster-autoscaler.yaml
    VAR=$(cat <<EOM
    apiVersion: kustomize.toolkit.fluxcd.io/v1beta1
    kind: Kustomization
    metadata:
      name: cluster-autoscaler
      namespace: flux-system
    spec:
      interval: 10m0s
      sourceRef:
        kind: GitRepository
        name: flux-system
      path: ./infrastructure/cluster-autoscaler
      prune: true
      validation: client
      patches:
      - target:
          group: helm.toolkit.fluxcd.io
          version: v2beta1
          kind: HelmRelease
          name: "cluster-autoscaler"
          namespace: "flux-system"
        patch: |-
          - op: replace
            path: "/spec/values/rbac/serviceAccount/annotations/eks.amazonaws.com~1role-arn"
            value: $ROLENAME
    EOM
    )
    echo "$VAR" > $FILE

    ROLENAME="arn:aws:iam::$AWS_ACCOUNTNUMBER:role/external-dns-service-account-$ENV_NAME"
    FILE=$REPO_PATH/clusters/$ENV_NAME/kustomization-external-dns.yaml
    VAR=$(cat <<EOM
    apiVersion: kustomize.toolkit.fluxcd.io/v1beta1
    kind: Kustomization
    metadata:
      name: external-dns
      namespace: flux-system
    spec:
      interval: 10m0s
      sourceRef:
        kind: GitRepository
        name: flux-system
      path: ./infrastructure/external-dns
      prune: true
      validation: client
      patches:
      - target:
          group: helm.toolkit.fluxcd.io
          version: v2beta1
          kind: HelmRelease
          name: "external-dns"
          namespace: "flux-system"
        patch: |-
          - op: replace
            path: "/spec/values/serviceAccount/annotations/eks.amazonaws.com~1role-arn"
            value: $ROLENAME
    EOM
    )
    echo "$VAR" > $FILE

    cd $REPO_PATH
    git config --global user.email "automation@bitso.com"
    git config --global user.name "Automation"
    git add $FILE -v
    git commit -m "Added annotation for IAM role"
    git push origin $BRANCH