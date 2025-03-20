package create

import (
	"archive/zip"
	"fmt"
	"io"
	"path/filepath"
	"strings"

	"os"

	"github.com/chainlaunch/chainlaunch/pkg/logger"
	"github.com/pkg/errors"
	"github.com/spf13/cobra"
)

func (c *createCmd) unzipFileInFolder(filePath string, dst string) error {
	archive, err := zip.OpenReader(filePath)
	if err != nil {
		return err
	}
	defer func(archive *zip.ReadCloser) {
		err := archive.Close()
		if err != nil {
			c.logger.Warnf("failed to close archive: %s", err)
		}
	}(archive)

	for _, f := range archive.File {
		filePath := filepath.Join(dst, f.Name)
		fmt.Println("unzipping file ", filePath)

		if !strings.HasPrefix(filePath, filepath.Clean(dst)+string(os.PathSeparator)) {
			fmt.Println("invalid file path")
			return fmt.Errorf("invalid file path: %s", filePath)
		}
		if f.FileInfo().IsDir() {
			fmt.Println("creating directory...")
			err = os.MkdirAll(filePath, os.ModePerm)
			if err != nil {
				return err
			}
			continue
		}

		if err := os.MkdirAll(filepath.Dir(filePath), os.ModePerm); err != nil {
			return err
		}

		dstFile, err := os.OpenFile(filePath, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, f.Mode())
		if err != nil {
			return err
		}

		fileInArchive, err := f.Open()
		if err != nil {
			return err
		}

		if _, err := io.Copy(dstFile, fileInArchive); err != nil {
			return err
		}

		err = dstFile.Close()
		if err != nil {
			return err
		}
		err = fileInArchive.Close()
		if err != nil {
			return err
		}
	}
	return nil
}

type createCmd struct {
	language string
	out      string
	name     string
	logger   *logger.Logger
}

func (c createCmd) run() error {
	var err error
	// get zip by language

	// create output folder, check it doesn't exist
	err = os.MkdirAll(c.out, 0755)
	if err != nil {
		return err
	}
	templateFilePath := "templates/" + c.language + ".zip"
	if _, err := os.Stat(templateFilePath); os.IsNotExist(err) {
		return errors.Wrapf(err, "failed to find template file %q", templateFilePath)
	}
	// unzip to output folder
	err = c.unzipFileInFolder(templateFilePath, c.out)
	if err != nil {
		return err
	}
	return nil
}
func NewCreateCmd(logger *logger.Logger) *cobra.Command {
	c := &createCmd{
		logger: logger,
	}
	cmd := &cobra.Command{
		Use:   "create",
		Short: "Create a new chaincode project",
		Long:  `Create a new chaincode project`,
		RunE: func(cmd *cobra.Command, args []string) error {
			return c.run()
		},
	}
	f := cmd.Flags()
	f.StringVarP(&c.out, "out", "o", "", "output directory")
	f.StringVarP(&c.language, "lang", "l", "", "language directory")
	f.StringVarP(&c.name, "name", "n", "", "name of the chaincode")

	return cmd
}
