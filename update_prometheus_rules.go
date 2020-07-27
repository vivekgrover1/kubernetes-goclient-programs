package main

import (
	"bytes"
	"context"
	"encoding/json"
	"flag"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"time"

	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/tools/clientcmd"
)

func createconfigmap(cmname string, namespace string, path string, clientset *kubernetes.Clientset) {

	var rules []string
	configMapData := make(map[string]string, 0)

	err := filepath.Walk(path_infra, func(path string, info os.FileInfo, err error) error {
		rules = append(rules, info.Name())
		return nil
	})

	if err != nil {
		panic(err)
	}

	for _, file := range rules {
		if strings.HasSuffix(file, ".rules") == true {
			content, err := ioutil.ReadFile(fmt.Sprintf("%s/%s", path, file))
			if err != nil {
				log.Fatal(err)
			}
			// Convert []byte to string and print to screen
			filecontent := string(content)
			configMapData[file] = filecontent

		}
	}


	configMap := &corev1.ConfigMap{
		TypeMeta: metav1.TypeMeta{
			Kind:       "ConfigMap",
			APIVersion: "v1",
		},
		ObjectMeta: metav1.ObjectMeta{
			Name:      cmname,
			Namespace: namespace,
		},
		Data: configMapData,
	}
	_, err_cm := clientset.CoreV1().ConfigMaps(namespace).Create(context.TODO(), configMap, metav1.CreateOptions{})
	if err_cm != nil {
		panic(err_cm.Error())
	}

}

func checkconfigmap(namespace string, cmname string, clientset *kubernetes.Clientset) string {

	out, err := clientset.CoreV1().ConfigMaps(namespace).List(context.TODO(), metav1.ListOptions{})
	if err != nil {
		panic(err.Error())
	}
	for _, d := range out.Items {
		if cmname == d.Name {
			return cmname
		}
	}
	return "none"

}

func deleteconfigmap(namespace string, cmname string, clientset *kubernetes.Clientset) {

	err := clientset.CoreV1().ConfigMaps(namespace).Delete(context.TODO(), cmname, &metav1.DeleteOptions{})
	if err != nil {
		panic(err.Error())
	}
}

func checkrulefiles(path string) {

	command := "promtool"
	arg0 := "check"
	arg1 := "rules"

	var infra_rules []string

	err := filepath.Walk(path, func(path string, info os.FileInfo, err error) error {
		rules = append(infra_rules, info.Name())
		return nil
	})

	if err != nil {
		panic(err)
	}

	for _, file := range rules {
		if strings.HasSuffix(file, ".rules") == true {

			cmd := exec.Command(command, arg0, arg1, fmt.Sprintf("%s/%s", path, file))
			cmd.Stderr = os.Stderr
			stdout, err := cmd.Output()

			if err != nil {
				fmt.Fprintln(os.Stderr)
				fmt.Println(err.Error())
				os.Exit(1)
			}

			fmt.Println(string(stdout))
		}
	}

}


func promrule(env string) {

	var kubeconfig *string

	kubeconfig = flag.String(fmt.Sprintf("kubeconfig-%s", env), fmt.Sprintf("/root/kubeconfig-%s", env), "absolute path to the kubeconfig file")

	flag.Parse()

	// use the current context in kubeconfig
	config, err := clientcmd.BuildConfigFromFlags("", *kubeconfig)
	if err != nil {
		panic(err.Error())
	}

	// create the clientset
	clientset, err := kubernetes.NewForConfig(config)
	if err != nil {
		panic(err.Error())
	}
	path := “/rules"
	namespace := "monitoring"
	cmname := "prometheus-rule"

	output := checkconfigmap(namespace, cmname, clientset)

	if output != "none" {
		deleteconfigmap(namespace, cmname, clientset)
	}
	createconfigmap(cmname, namespace, path_infra, path_app, clientset)

}

func main() {

	path := "rules"
	checkrulefiles(path)
	promrule("prd")
		

}
