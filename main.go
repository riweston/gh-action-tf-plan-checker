/*
Copyright © 2023 NAME HERE <EMAIL ADDRESS>
*/
package main

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/hashicorp/terraform-exec/tfexec"
	tfjson "github.com/hashicorp/terraform-json"
	"github.com/urfave/cli/v2"
	"log"
	"os"
)

func main() {
	app := &cli.App{
		Flags: []cli.Flag{
			&cli.StringFlag{
				Name:     "tfplan",
				Aliases:  []string{"p"},
				Usage:    "Load terraform plan from `FILE`",
				Required: true,
			},
			&cli.BoolFlag{
				Name:    "allow-delete",
				Aliases: []string{"d"},
				Usage:   "Allow delete actions",
				Value:   false,
			},
			&cli.BoolFlag{
				Name:    "allow-update",
				Aliases: []string{"u"},
				Usage:   "Allow update actions",
				Value:   false,
			},
			&cli.BoolFlag{
				Name:    "allow-create",
				Aliases: []string{"c"},
				Usage:   "Allow create actions",
				Value:   false,
			},
		},
		Action: func(cCtx *cli.Context) error {
			if cCtx.String("tfplan") != "" {

				plan, err := ShowPlan(cCtx.String("tfplan"))
				if err != nil {
					log.Fatal(err)
				}
				_, err = json.Marshal(plan)
				if err != nil {
					log.Fatal(err)
				}
				var errFlags []string
				for _, flag := range []string{"allow-delete", "allow-update", "allow-create"} {
					if cCtx.Bool(flag) == false {
						check := CheckForResourceChange(plan, flag)
						if check != nil {
							errFlags = append(errFlags, check.Error())
						}
					}
				}
				return fmt.Errorf("Deny Changes: %v", errFlags)
			}
			return nil
		},
	}

	if err := app.Run(os.Args); err != nil {
		log.Fatal(err)
	}

}

func CheckForResourceChange(plan *tfjson.Plan, action string) (err error) {
	action = action[6:]
	for _, i := range plan.ResourceChanges {
		for _, j := range i.Change.Actions {
			if string(j) == action {
				err = fmt.Errorf("%s", action)
			}
		}
	}
	return err
}

func ShowPlan(p string) (*tfjson.Plan, error) {
	ctx := context.Background()
	pwd, err := os.Getwd()
	if err != nil {
		panic(err)
	}
	fileExists := testFile(pwd, p)
	if fileExists != nil {
		return nil, fileExists
	}
	tf, _ := tfexec.NewTerraform(pwd, "terraform")
	return tf.ShowPlanFile(ctx, p)
}

func testFile(pwd string, filename string) (err error) {
	_, err = os.Stat(fmt.Sprintf("%s/%s", pwd, filename))
	if os.IsNotExist(err) {
		return fmt.Errorf("No such file %s", filename)
	}
	return
}
