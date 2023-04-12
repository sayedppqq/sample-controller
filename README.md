# sample-controller #

Group Name: sayedppqq.dev

Version Name: v1alpha1

Resource Name: Kluster

## Code Gen ##
Code gen is needed for generating:
- DeepCopyObject
- Clientset
- Informer
- Lister

### Procedure ###

- import `"k8s.io/code-generator"` into `main.go`
- run `go mod tidy;go mod vendor`
- run `chmod +x ..../sample-controller/hack/codegen.sh`
- run `chmod +x vendor/k8s.io/code-generator/generate-groups.sh`
- run `hack/codegen.sh`
- again run `go mod tidy;go mod vendor`

#### To generate controller-gen ####
- Run `depelopmentDir=$(pwd)`

- Then run `controller-gen rbac:roleName=controller-perms crd paths=github.com/sayedppqq/sample-controller/pkg/apis/sayedppqq.dev/v1alpha1 crd:crdVersions=v1 output:crd:dir=$depelopmentDir/manifests output:stdout`

## Deploy custom resource ##

Just create a yaml file like `manifests/kluster.yaml` and apply.

Run `kubectl get Kluster`

## Resource ##
https://www.linkedin.com/pulse/kubernetes-custom-controllers-part-1-kritik-sachdeva/

https://www.linkedin.com/pulse/kubernetes-custom-controller-part-2-kritik-sachdeva/

https://github.com/ishtiaqhimel/crd-controller/tree/master
