package cmd

import (
	"context"
	"fmt"
	"io"
	"log"
	"os"
	stdpath "path"
	"path/filepath"

	"github.com/spf13/cobra"
	"go-hep.org/x/hep/xrootd"
	"go-hep.org/x/hep/xrootd/xrdfs"
	"go-hep.org/x/hep/xrootd/xrdio"
)

func init() {
	rootCmd.AddCommand(downloadCmd)
	downloadCmd.PersistentFlags().StringVarP(&outdir, "out", "o", "download", "Output directoy")

}

var (
	outdir string

	downloadCmd = &cobra.Command{
		Use:   "download",
		Short: "Download files belonging to a record",
		Long: `This command will download all files associated
with a record.`,
		Run: func(cmd *cobra.Command, args []string) {
			err := verifyRecordID()
			if err != nil {
				er(err)
			}
			recordJSON, err := getRecordJSON()
			if err != nil {
				er(err)
			}
			filesList, err := getFilesList(recordJSON, true)
			if err != nil {
				er(err)
			}
			downloadFiles(filesList)
			if err != nil {
				er(err)
			}
		},
	}
)

func downloadFiles(filesList []string) error {

	recordOutDir := filepath.Join(outdir, recordID)
	err := os.MkdirAll(recordOutDir, 0755)
	if err != nil {
		return err
	}
	for i := range filesList {
		destPath := filepath.Join(recordOutDir, stdpath.Base(filesList[i]))
		fmt.Println(filesList[i], "-->", destPath)
		err = xrdcopy(destPath, filesList[i], true)
		if err != nil {
			return err
		}
		break
	}

	return err
}

func xrdcopy(dst, srcPath string, verbose bool) error {
	cli, src, err := xrdremote(srcPath)
	if err != nil {
		return err
	}
	defer cli.Close()

	ctx := context.Background()

	fs := cli.FS()
	var jobs jobs
	var addDir func(root, src string) error

	addDir = func(root, src string) error {
		fi, err := fs.Stat(ctx, src)
		if err != nil {
			return fmt.Errorf("could not stat remote src: %w", err)
		}
		switch {
		case fi.IsDir():
			dst := stdpath.Join(root, stdpath.Base(src))
			err = os.MkdirAll(dst, 0755)
			if err != nil {
				return fmt.Errorf("could not create output directory: %w", err)
			}

			ents, err := fs.Dirlist(ctx, src)
			if err != nil {
				return fmt.Errorf("could not list directory: %w", err)
			}
			for _, e := range ents {
				err = addDir(dst, stdpath.Join(src, e.Name()))
				if err != nil {
					return err
				}
			}
		default:
			jobs.add(job{
				fs:  fs,
				src: src,
				dst: stdpath.Join(root, stdpath.Base(src)),
			})
		}
		return nil
	}

	fiSrc, err := fs.Stat(ctx, src)
	if err != nil {
		return fmt.Errorf("could not stat remote src: %w", err)
	}

	fiDst, errDst := os.Stat(dst)
	switch {
	case fiSrc.IsDir():
		switch {
		case errDst != nil && os.IsNotExist(errDst):
			err = os.MkdirAll(dst, 0755)
			if err != nil {
				return fmt.Errorf("could not create output directory: %w", err)
			}
			ents, err := fs.Dirlist(ctx, src)
			if err != nil {
				return fmt.Errorf("could not list directory: %w", err)
			}
			for _, e := range ents {
				err = addDir(dst, stdpath.Join(src, e.Name()))
				if err != nil {
					return err
				}
			}

		case errDst != nil:
			return fmt.Errorf("could not stat local dst: %w", errDst)
		case fiDst.IsDir():
			err = addDir(dst, src)
			if err != nil {
				return err
			}
		}

	default:
		switch {
		case errDst != nil && os.IsNotExist(errDst):
			// ok... dst will be the output file.
		case errDst != nil:
			return fmt.Errorf("could not stat local dst: %w", errDst)
		case fiDst.IsDir():
			dst = stdpath.Join(dst, stdpath.Base(src))
		}

		jobs.add(job{
			fs:  fs,
			src: src,
			dst: dst,
		})
	}

	n, err := jobs.run(ctx)
	if verbose {
		log.Printf("transferred %d bytes", n)
	}
	return err
}

func xrdremote(name string) (client *xrootd.Client, path string, err error) {
	url, err := xrdio.Parse(name)
	if err != nil {
		return nil, "", fmt.Errorf("could not parse %q: %w", name, err)
	}

	path = url.Path
	client, err = xrootd.NewClient(context.Background(), url.Addr, url.User)
	return client, path, err
}

type job struct {
	fs  xrdfs.FileSystem
	src string
	dst string
}

func (j job) run(ctx context.Context) (int, error) {
	var (
		o   io.WriteCloser
		err error
	)
	switch j.dst {
	case "-", "":
		o = os.Stdout
	case ".":
		j.dst = stdpath.Base(j.src)
		fallthrough
	default:
		o, err = os.Create(j.dst)
		if err != nil {
			return 0, fmt.Errorf("could not create output file: %w", err)
		}
	}
	defer o.Close()

	f, err := xrdio.OpenFrom(j.fs, j.src)
	if err != nil {
		return 0, err
	}
	defer f.Close()

	n, err := io.CopyBuffer(o, f, make([]byte, 16*1024*1024))
	if err != nil {
		return int(n), fmt.Errorf("could not copy to output file: %w", err)
	}

	err = o.Close()
	if err != nil {
		return int(n), fmt.Errorf("could not close output file: %w", err)
	}

	return int(n), nil
}

type jobs struct {
	slice []job
}

func (js *jobs) add(j job) {
	js.slice = append(js.slice, j)
}

func (js *jobs) run(ctx context.Context) (int, error) {
	var n int
	for _, j := range js.slice {
		nn, err := j.run(ctx)
		n += nn
		if err != nil {
			return n, err
		}
	}
	return n, nil
}
