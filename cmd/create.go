package cmd

import (
	"bytes"
	"flag"
	"fmt"
	"html/template"
	"log"
	"os"
	"time"

	"github.com/bendahl/tnt/cmd/util"
)

func Create(args []string) {
	var requestsMem string
	var requestsCPU uint
	var limitsMem string
	var limitsCPU uint
	var namespace string
	var team string
	var duration time.Duration

	err := parseArgs(args, &requestsMem, &requestsCPU, &limitsMem, &limitsCPU, &namespace, &team, &duration)
	if err != nil {
		log.Fatalf("failed to parse arguments: %v\n", err)
	}
	err = createNamespace(team, namespace)
	if err != nil {
		log.Fatalf("failed to create team namespace: %v", err)
	}
	err = createResourceQuota(team, namespace, requestsCPU, requestsMem, limitsCPU, limitsMem)
	if err != nil {
		log.Fatalf("failed to create resource quota: %v", err)
	}
	adminName := team + "-admin"
	err = createServiceAccount(adminName, team)
	if err != nil {
		log.Fatalf("failed to create service account for team admin: %v", err)
	}
	err = createRoleBinding(adminName, namespace, adminName, team)
	if err != nil {
		log.Fatalf("failed to create rolebinding: %v", err)
	}
	token, err := util.Kubectl(fmt.Sprintf("create token %s -n kube-system --duration %s", adminName, duration))
	if err != nil {
		log.Fatalf("failed to obtain serviceaccount token: %v", err)
	}
	kubeconfig, err := createKubeconfig(namespace, adminName, token)
	if err != nil {
		log.Fatalf("failed to create kubeconfig: %v", err)
	}
	fmt.Println(kubeconfig)
}

const roleBindingTemplate = `
apiVersion: rbac.authorization.k8s.io/v1
kind: RoleBinding
metadata:
  name: {{.RoleBindingName}}
  namespace: {{.Namespace}}
  labels:
    team: {{.Team}}
roleRef:
  apiGroup: rbac.authorization.k8s.io
  kind: ClusterRole
  name: admin
subjects:
- kind: ServiceAccount
  name: {{.ServiceAccountName}}
  namespace: kube-system
`

func createRoleBinding(roleBindingName, namespace, serviceAccountName, team string) error {
	params := struct {
		RoleBindingName    string
		Namespace          string
		ServiceAccountName string
		Team               string
	}{
		RoleBindingName:    roleBindingName,
		Namespace:          namespace,
		ServiceAccountName: serviceAccountName,
		Team:               team,
	}
	roleBinding, err := renderTemplate(roleBindingTemplate, params)
	if err != nil {
		return err
	}
	err = applyManifest(serviceAccountName+"_rolebinding.yml", roleBinding)
	if err != nil {
		return err
	}
	return nil
}

const kubeconfigTemplate = `
apiVersion: v1
clusters:
- cluster:
    certificate-authority-data: {{.CaCert}}
    server: {{.ServerAddress}}
  name: {{.ClusterName}}
contexts:
- context:
    cluster: {{.ClusterName}}
    namespace: {{.Namespace}}
    user: {{.UserName}}
  name: {{.ClusterName}}
current-context: {{.ClusterName}}
kind: Config
users:
- name: {{.UserName}}
  user:
    token: {{.Token}}
`

func createKubeconfig(namespace, userName, token string) (string, error) {
	serverUrl, err := util.Kubectl("config view -o jsonpath='{.clusters[0].cluster.server}'")
	if err != nil {
		return "", err
	}
	clusterName, err := util.Kubectl("config view -o jsonpath='{.clusters[0].name}'")
	if err != nil {
		return "", err
	}
	caCert, err := util.Kubectl("config view --raw -o jsonpath='{.clusters[0].cluster.certificate-authority-data}'")
	if err != nil {
		return "", err
	}
	params := struct {
		ServerAddress string
		CaCert        string
		ClusterName   string
		Namespace     string
		UserName      string
		Token         string
	}{
		ServerAddress: serverUrl,
		CaCert:        caCert,
		ClusterName:   clusterName,
		Namespace:     namespace,
		UserName:      userName,
		Token:         token,
	}
	kubeconfig, err := renderTemplate(kubeconfigTemplate, params)
	if err != nil {
		return "", err
	}
	return kubeconfig, nil
}

func applyManifest(manifestName string, manifest string) error {
	tmp, err := os.CreateTemp("", manifestName)
	if err != nil {
		return fmt.Errorf("failed to create temp file for manifest: %v\n", err)
	}
	defer tmp.Close()
	defer os.Remove(tmp.Name())
	_, err = tmp.WriteString(manifest)
	if err != nil {
		return fmt.Errorf("Failed to write manifest to temp file: %v\n", err)
	}
	res, err := util.Kubectl("apply -f " + tmp.Name())
	if err != nil {
		return fmt.Errorf("failed to apply manifest: %v, output: %s", err, res)
	}
	return nil
}

func createResourceQuota(team string, namespace string, requestsCPU uint, requestsMemory string, limitsCPU uint,
	limitsMemory string) error {
	params := struct {
		Team           string
		Namespace      string
		RequestsCPU    uint
		RequestsMemory string
		LimitsCPU      uint
		LimitsMemory   string
	}{
		Team:           team,
		Namespace:      namespace,
		RequestsCPU:    requestsCPU,
		RequestsMemory: requestsMemory,
		LimitsCPU:      limitsCPU,
		LimitsMemory:   limitsMemory,
	}
	quota, err := renderTemplate(quotaTemplate, params)
	if err != nil {
		return fmt.Errorf("failed to create resource quota: %v", err)
	}
	return applyManifest(team+"_quota.yml", quota)
}

func createNamespace(team string, namespace string) error {
	params := struct {
		Team      string
		Namespace string
	}{Team: team, Namespace: namespace}
	namespaceManifest, err := renderTemplate(namespaceTemplate, params)
	if err != nil {
		return fmt.Errorf("failed to create namespace manifest: %v\n", err)
	}
	return applyManifest(team+"_namespace.yml", namespaceManifest)
}

func createServiceAccount(saName, team string) error {
	params := struct {
		ServiceAccountName string
		Team               string
	}{
		ServiceAccountName: saName,
		Team:               team,
	}
	saManifest, err := renderTemplate(serviceAccountTemplate, params)
	if err != nil {
		return fmt.Errorf("failed to create serviceaccount manifest: %v\n", err)
	}
	return applyManifest(saName+"_serviceaccount.yml", saManifest)
}

func parseArgs(args []string, requestsMem *string, requestsCPU *uint, limitsMem *string, limitsCPU *uint, namespace *string, team *string, duration *time.Duration) error {
	cfs := flag.NewFlagSet("create", flag.ExitOnError)
	cfs.StringVar(requestsMem, "requests-mem", "", "requests memory limit (e.g. '1Gi')")
	cfs.UintVar(requestsCPU, "requests-cpu", 0, "requests CPU limit (e.g. 1)")
	cfs.StringVar(limitsMem, "limits-mem", "", "limits memory limit (e.g. '1Gi')")
	cfs.UintVar(limitsCPU, "limits-cpu", 0, "limits CPU limit (e.g. 1)")
	cfs.StringVar(namespace, "n", "", "name of the team namespace")
	cfs.StringVar(team, "t", "", "name of the team")
	cfs.DurationVar(duration, "d", 48*time.Hour, "validity duration of sa token (e.g. 48h)")

	err := cfs.Parse(args)
	if err != nil {
		return fmt.Errorf("failed to parse arguments: %v", err)
	}

	if *namespace == "" {
		return fmt.Errorf("no namespace provided")
	}
	if *team == "" {
		return fmt.Errorf("no team name given")
	}
	return nil
}

const namespaceTemplate = `
apiVersion: v1
kind: Namespace
metadata:
  name: {{if .Namespace -}} {{.Namespace}} {{else -}} {{.Team}} {{end}}
  labels:
    team: {{.Team}}
spec: {}
`

func renderTemplate(templateStr string, params any) (string, error) {
	tpl := template.Must(template.New("tpl").Parse(templateStr))
	var buff bytes.Buffer
	err := tpl.Execute(&buff, params)
	return buff.String(), err
}

const quotaTemplate = `
apiVersion: v1
kind: ResourceQuota
metadata:
  name: {{.Team}}-quota
  namespace: {{if .Namespace -}} {{.Namespace}} {{else -}} {{.Team}} {{end}}
  labels:
    team: {{.Team}}
spec:
  hard:
	{{- if (gt .RequestsCPU 0) }}
    requests.cpu: {{.RequestsCPU}}{{ end -}}
	{{if .RequestsMemory }}
    requests.memory: {{.RequestsMemory}}{{ end -}}
	{{if (gt .LimitsCPU 0)}}
    limits.cpu: {{.LimitsCPU}}{{ end -}}
	{{if .LimitsMemory}}
    limits.memory: {{.LimitsMemory}}{{ end -}}
`

const serviceAccountTemplate = `
apiVersion: v1
kind: ServiceAccount
metadata:
  name: {{.ServiceAccountName}}
  namespace: kube-system
  labels:
    team: {{.Team}}

`
