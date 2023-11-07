// SPDX-License-Identifier: Apache-2.0

package root

import (
	"github.com/gittuf/gittuf/internal/cmd/clone"
	"github.com/gittuf/gittuf/internal/cmd/policy"
	"github.com/gittuf/gittuf/internal/cmd/rsl"
	"github.com/gittuf/gittuf/internal/cmd/trust"
	"github.com/gittuf/gittuf/internal/cmd/verifycommit"
	"github.com/gittuf/gittuf/internal/cmd/verifyref"
	"github.com/gittuf/gittuf/internal/cmd/verifytag"
	"github.com/gittuf/gittuf/internal/cmd/version"
	"github.com/gittuf/gittuf/internal/logging"
	"github.com/spf13/cobra"
)

func addRootFlags(cmd *cobra.Command) {
	cmd.PersistentFlags().BoolP("verbose", "v", false, "Enable verbose logging")
}

func New() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "gittuf",
		Short: "A security layer for Git repositories, powered by TUF",
		PersistentPreRunE: func(cmd *cobra.Command, args []string) error {
			isVerbose, err := cmd.Flags().GetBool("verbose")
			if err != nil {
				return err
			}

			err = logging.InitLogger(isVerbose)
			if err != nil {
				return err
			}

			return nil
		},
	}
	addRootFlags(cmd)
	cmd.AddCommand(clone.New())
	cmd.AddCommand(trust.New())
	cmd.AddCommand(policy.New())
	cmd.AddCommand(rsl.New())
	cmd.AddCommand(verifycommit.New())
	cmd.AddCommand(verifyref.New())
	cmd.AddCommand(verifytag.New())
	cmd.AddCommand(version.New())

	return cmd
}
