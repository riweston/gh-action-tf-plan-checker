/*
Copyright © 2023 Richard Weston github@riweston.io
*/
package main

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/hashicorp/go-version"
	"github.com/hashicorp/hc-install/product"
	"github.com/hashicorp/hc-install/releases"
	"github.com/hashicorp/terraform-exec/tfexec"
	tfjson "github.com/hashicorp/terraform-json"
	"github.com/pkg/errors"
	"github.com/urfave/cli/v2"
	"log"
	"os"
	"os/exec"
	"path/filepath"
)

var terraformBinPath string
var terraformPlan *tfjson.Plan

/*
terraformPlanBin is a bool to check if the plan is a binary file

The code that runs this is WIP and will be updated in the future
as of today this should be disabled.
*/
var terraformPlanBin = false

func main() {
	app := &cli.App{
		Name:  "tfplan-check",
		Usage: "Check terraform plan for actions",
		Flags: []cli.Flag{
			&cli.StringFlag{
				Name:     "terraform-plan",
				Aliases:  []string{"p"},
				Usage:    "Load terraform plan from `FILE`",
				Required: true,
			},
			&cli.BoolFlag{
				Name:    "allow-delete",
				Aliases: []string{"d"},
				Usage:   "Allow delete actions",
				Value:   false,
				EnvVars: []string{"TFPLAN_CHECK_ALLOW_DELETE"},
			},
			&cli.BoolFlag{
				Name:    "allow-update",
				Aliases: []string{"u"},
				Usage:   "Allow update actions",
				Value:   false,
				EnvVars: []string{"TFPLAN_CHECK_ALLOW_UPDATE"},
			},
			&cli.BoolFlag{
				Name:    "allow-create",
				Aliases: []string{"c"},
				Usage:   "Allow create actions",
				Value:   false,
				EnvVars: []string{"TFPLAN_CHECK_ALLOW_CREATE"},
			},
		},
		Action: func(cCtx *cli.Context) error {
			if cCtx.String("terraform-plan") != "" {
				checkPlanType(cCtx.String("terraform-plan"))

				var errFlags []string
				for _, flag := range []string{"allow-delete", "allow-update", "allow-create"} {
					if cCtx.Bool(flag) == false {
						check := CheckForResourceChange(terraformPlan, flag)
						if check != nil {
							errFlags = append(errFlags, check.Error())
						}
					}
				}
				if len(errFlags) > 0 {
					log.Fatalf("The following actions are not allowed: %s", errFlags)
				}
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

func checkPlanType(plan string) {
	planPath := testFile(plan)
	showPlanJson(planPath)
	if terraformPlan == nil && terraformPlanBin == true {
		getTerraformPath()
		showPlan(planPath)
	} else {
		showPlanJson(planPath)
	}
}

func showPlanJsonE(plan string) error {
	planBytes, err := os.ReadFile(plan)
	if err != nil {
		return fmt.Errorf("error reading plan file: %s", err)
	}
	if json.Valid(planBytes) {
		return json.Unmarshal(planBytes, &terraformPlan)
	} else {
		return fmt.Errorf("plan is not a valid json file")
	}
}

func showPlanJson(plan string) {
	err := showPlanJsonE(plan)
	if err != nil {
		log.Fatalf("error running showPlan: %s", err)
	}
}

func showPlanE(ctx context.Context, plan string) (*tfjson.Plan, error) {
	tf := initTerraform(ctx, filepath.Dir(plan))
	options := []tfexec.ShowOption{}
	planTmp, err := tf.ShowPlanFile(ctx, plan, options...)
	var mismatch *tfexec.ErrVersionMismatch
	if errors.As(err, &mismatch) {
		installTerraformVersion(mismatch.MinInclusive)
		//showPlanE(ctx, plan)
		return nil, mismatch
	}
	return planTmp, err
}

func showPlan(plan string) {
	ctx := context.Background()
	var err error
	terraformPlan, err = showPlanE(ctx, plan)
	if err != nil {
		log.Fatalf("error running ShowPlan: %s", err)
	}
}

func newTerraformE(p string) (tf *tfexec.Terraform, err error) {
	return tfexec.NewTerraform(p, terraformBinPath)
}

func newTerraform(p string) *tfexec.Terraform {
	tf, err := newTerraformE(p)
	if err != nil {
		log.Fatalf("error getting Terraform: %s", err)
	}
	return tf
}

func initTerraformE(ctx context.Context, p string) (tf *tfexec.Terraform, err error) {
	tf = newTerraform(p)
	err = tf.Init(ctx, tfexec.Upgrade(true))
	return
}

func initTerraform(ctx context.Context, p string) *tfexec.Terraform {
	tf, err := initTerraformE(ctx, p)
	if err != nil {
		log.Fatalf("error running Init: %s", err)
	}
	return tf
}

func testFileE(plan string) (planPath string, err error) {
	planPath, err = filepath.Abs(plan)
	if err != nil {
		return "", fmt.Errorf("error getting absolute path: %s", err)
	}
	return
}

func testFile(plan string) string {
	planPath, err := testFileE(plan)
	if err != nil {
		log.Fatalf("error getting plan file: %s", err)
	}
	return planPath
}

func getTerraformPath() (path string) {
	var err error
	terraformBinPath, err = exec.LookPath("terraform")
	//if err == nil {
	// Flip this back after testing
	if err != nil {
		installTerraformLatest()
	}
	return path
}

func installTerraformVersionE(v string) (string, error) {
	installer := &releases.ExactVersion{
		Product: product.Terraform,
		Version: version.Must(version.NewVersion(v)),
	}
	return installer.Install(context.Background())
}

func installTerraformVersion(v string) {
	var err error
	terraformBinPath, err = installTerraformVersionE(v)
	if err != nil {
		log.Fatalf("error installing Terraform: %s", err)
	}
}

func installTerraformLatestE() (string, error) {
	installer := &releases.LatestVersion{
		Product: product.Terraform,
	}
	return installer.Install(context.Background())
}

func installTerraformLatest() {
	var err error
	terraformBinPath, err = installTerraformLatestE()
	if err != nil {
		log.Fatalf("error installing Terraform: %s", err)
	}
}
