package main

import (
	"fmt"
	"strings"

	"github.com/alecthomas/kingpin"
	"github.com/gosuri/uitable"
	"k8s.io/client-go/rest"

	"github.com/previousnext/client"
)

func main() {
	kingpin.Parse()

	config, err := rest.InClusterConfig()
	if err != nil {
		panic(err.Error())
	}

	efsClient, err := client.NewClient(config)
	if err != nil {
		panic(err)
	}

	all, err := efsClient.ListAll()
	if err != nil {
		panic(err)
	}

	table := uitable.New()
	table.MaxColWidth = 80

	table.AddRow("NAMESPACE", "NAME", "REGION", "SUBNETS", "SECURITY", "ID", "CYCLE", "LAST UPDATE")
	for _, fs := range all.Items {
		table.AddRow(fs.Metadata.Namespace, fs.Metadata.Name, fs.Spec.Region, strings.Join(fs.Spec.Subnets, ", "), fs.Spec.SecurityGroup, fs.Status.ID, fs.Status.LifeCycleState, fs.Status.LastUpdate)
	}
	fmt.Println(table)
}
