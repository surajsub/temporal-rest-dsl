package utils

import (
	"fmt"
	"log"
	"os"
	"strings"
)

// WriteTerraformVariablesFile writes the given parameters to a vars.tf.old file in the specified directory
func WriteTerraformVariablesFile(dir string, parameters map[string]string) error {
	filePath := fmt.Sprintf("%s/vars.tf.old", dir)
	log.Printf("Writing variables to %s", filePath)

	var tfVarsContent strings.Builder
	for key, value := range parameters {
		tfVarsContent.WriteString(fmt.Sprintf("variable \"%s\" {\n  default = \"%s\"\n}\n\n", key, value))
	}

	if err := os.WriteFile(filePath, []byte(tfVarsContent.String()), 0644); err != nil {
		return fmt.Errorf("failed to write vars.tf.old file: %w", err)
	}

	return nil
}
