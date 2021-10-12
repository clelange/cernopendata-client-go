package cmd

import (
	"context"
	"fmt"
	"io"
	"log"
	"os"
	stdpath "path"
	"path/filepath"
	"time"

	"github.com/cavaliercoder/grab"
	"github.com/spf13/cobra"
	"go-hep.org/x/hep/xrootd"
	"go-hep.org/x/hep/xrootd/xrdfs"
	"go-hep.org/x/hep/xrootd/xrdio"
)

func init() {
	rootCmd.AddCommand(downloadCmd)
	downloadCmd.PersistentFlags().StringVarP(&outdir, "out", "o", "download", "Output directory")
	downloadCmd.PersistentFlags().StringVarP(&protocol, "protocol", "p", "http", "Protocol to be used (http or root)")
	downloadCmd.PersistentFlags().IntVar(&parallelDownloads, "parallel", 5, "Number of parallel downloads (http only)")

}

var (
	outdir            string
	parallelDownloads int

	downloadCmd = &cobra.Command{
		Use:   "download",
		Short: "Download files belonging to a record",
		Long: `This command will download all files associated
with a record.`,
		Run: func(cmd *cobra.Command, args []string) {
			if err := validateProtocolChoice(); err != nil {
				er(err)
			}
			if err := verifyUniqueID(); err != nil {
				er(err)
			}
			recordJSON, err := getRecordJSON()
			if err != nil {
				er(err)
			}
			filesList, err := getFilesList(recordJSON)
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

	if protocol == "root" {
		for i := range filesList {
			destPath := filepath.Join(recordOutDir, stdpath.Base(filesList[i]))
			fmt.Println(filesList[i], "-->", destPath)
			err = xrdcopy(destPath, filesList[i], true)
			if err != nil {
				return err
			}
		}
	} else { // protocol == "http"
		err = grabFiles(filesList, recordOutDir)
		if err != nil {
			return err
		}
	}

	return err
}

func grabFiles(filesList []string, recordOutDir string) error {

	fmt.Printf("Downloading %d files using %d parallel threads...\n", len(filesList), parallelDownloads)
	respch, err := grab.GetBatch(parallelDownloads, recordOutDir, filesList...)
	if err != nil {
		return err
	}

	// start a ticker to update progress every 1000ms
	t := time.NewTicker(1000 * time.Millisecond)

	// monitor downloads
	completed := 0
	inProgress := 0
	responses := make([]*grab.Response, 0)
	for completed < len(filesList) {
		select {
		case resp := <-respch:
			// a new response has been received and has started downloading
			// (nil is received once, when the channel is closed by grab)
			if resp != nil {
				responses = append(responses, resp)
			}

		case <-t.C:
			// clear lines
			if inProgress > 0 {
				fmt.Printf("\033[%dA\033[K", inProgress)
			}

			// update completed downloads
			for i, resp := range responses {
				if resp != nil && resp.IsComplete() {
					// print final result
					if resp.Err() != nil {
						fmt.Fprintf(os.Stderr, "Error downloading %s: %v\n", resp.Request.URL(), resp.Err())
					} else {
						fmt.Printf("Finished %s %d / %d bytes (%d%%)\n", resp.Filename, resp.BytesComplete(), resp.Size, int(100*resp.Progress()))
					}

					// mark completed
					responses[i] = nil
					completed++
				}
			}

			// update downloads in progress
			inProgress = 0
			for _, resp := range responses {
				if resp != nil {
					inProgress++
					fmt.Printf("Downloading %s %d / %d bytes (%d%%)\033[K\n", resp.Filename, resp.BytesComplete(), resp.Size, int(100*resp.Progress()))
				}
			}
		}
	}

	t.Stop()

	fmt.Printf("%d files successfully downloaded.\n", len(filesList))

	return nil

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
