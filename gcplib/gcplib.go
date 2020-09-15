package gcplib

import (
	"bufio"
	"context"
	"errors"
	"io"
	"io/ioutil"
	"os"
	"sync"

	"cloud.google.com/go/storage"
	"github.com/cafecoder-dev/cafecoder-container-client/types"
	"github.com/cafecoder-dev/cafecoder-judge/src/sqllib"
	"google.golang.org/api/option"
)

const credentialFilePath = "./key.json"

var client *storage.Client
var once sync.Once

func DownloadSourceCode(ctx context.Context, path string, name string) error {
	once.Do(func() {
		c, _ := storage.NewClient(ctx, option.WithCredentialsFile(credentialFilePath))
		client = c
	})

	fp, err := os.Create(name)
	if err != nil {
		return err
	}

	bucket := "cafecoder-submit-source"
	obj := client.Bucket(bucket).Object(path)
	reader, err := obj.NewReader(ctx)
	if err != nil {
		return err
	}
	defer reader.Close()

	tee := io.TeeReader(reader, fp)
	s := bufio.NewScanner(tee)
	for s.Scan() {
	}
	if err := s.Err(); err != nil {
		return err
	}

	return nil
}

func DownloadTestcase(ctx context.Context, problemUUID string, testcaseName string) ([]byte, []byte, error) {
	once.Do(func() {
		c, _ := storage.NewClient(ctx, option.WithCredentialsFile(credentialFilePath))
		client = c
	})

	bucketName := "cafecoder-testcase"
	bucket := client.Bucket(bucketName)

	reader, err := bucket.Object(problemUUID + "/input/" + testcaseName).NewReader(ctx)
	if err != nil {
		return nil, nil, err
	}
	defer reader.Close()

	inputData, err := ioutil.ReadAll(reader)
	if err != nil {
		return nil, nil, err
	}

	reader, err = bucket.Object(problemUUID + "/output/" + testcaseName).NewReader(ctx)
	if err != nil {
		return nil, nil, err
	}
	defer reader.Close()

	outputData, err := ioutil.ReadAll(reader)
	if err != nil {
		return nil, nil, err
	}

	return inputData, outputData, nil
}

func GetTestcases(ctx context.Context, problemID string) ([]types.TestcaseGORM, error) {
	var (
		testcases []types.TestcaseGORM
		problem   types.ProblemsGORM
		err       error
	)

	db, err := sqllib.NewDB()
	if err != nil {
		return nil, err
	}
	defer db.Close()

	db.
		Table("problems").
		Where("id = ? AND deleted_at IS NULL", problemID).
		First(&problem)

	db.
		Table("testcases").
		Where("problem_id=? AND deleted_at IS NULL", problemID).
		Order("id").
		Find(&testcases)

	if len(testcases) == 0 {
		return nil, errors.New("not found testcases")
	}

	for i := 0 ; i < len(testcases) ; i++ {
		testcases[i].Input, testcases[i].Output, err = DownloadTestcase(ctx, problem.UUID, testcases[i].Name)
		if err != nil {
			return nil, err
		}
	}

	return testcases, nil
}
