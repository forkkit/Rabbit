// Copyright 2019 Clivern. All rights reserved.
// Use of this source code is governed by the MIT
// license that can be found in the LICENSE file.

package controller

import (
	"fmt"
	"net/http"
	"strings"

	"github.com/clivern/hippo"
	"github.com/clivern/rabbit/internal/app/model"
	"github.com/clivern/rabbit/pkg"
	"github.com/gin-gonic/gin"
	"github.com/spf13/viper"
	"go.uber.org/zap"
)

// GitlabListener controller
func GitlabListener(c *gin.Context, messages chan<- string) {
	rawBody, _ := c.GetRawData()
	body := string(rawBody)

	parser := &pkg.GitlabWebhookParser{
		GitlabEvent: c.GetHeader("X-Gitlab-Event"),
		GitlabToken: c.GetHeader("X-Gitlab-Token"),
		Body:        body,
	}

	logger, _ := hippo.NewLogger(
		viper.GetString("log.level"),
		viper.GetString("log.format"),
		[]string{viper.GetString("log.output")},
	)

	defer logger.Sync()

	ok := parser.VerifySecret(viper.GetString("integrations.gitlab.webhook_secret"))

	if !ok && viper.GetString("integrations.gitlab.webhook_secret") != "" {
		c.JSON(http.StatusForbidden, gin.H{
			"status": "Oops!",
		})
		return
	}

	pushEvent := pkg.GitlabTagPushEvent{}
	ok, err := pushEvent.LoadFromJSON(rawBody)

	if err != nil {
		logger.Info(fmt.Sprintf(
			`Invalid event received %s`,
			body,
		), zap.String("CorrelationID", c.Request.Header.Get("X-Correlation-ID")))

		c.JSON(http.StatusForbidden, gin.H{
			"status": "Oops!",
		})
		return
	}

	if !ok {
		c.JSON(http.StatusForbidden, gin.H{
			"status": "Oops!",
		})
		return
	}

	if pushEvent.TotalCommitsCount <= 0 || pushEvent.EventName != "tag_push" || parser.GetGitlabEvent() != "Tag Push Hook" {
		c.JSON(http.StatusOK, gin.H{
			"status": "Nice, Skip!",
		})
		return
	}

	// Push event received
	href := strings.ReplaceAll(
		viper.GetString("integrations.gitlab.https_format"),
		"[.RepoFullName]",
		pushEvent.Project.PathWithNamespace,
	)

	if viper.GetString("integrations.gitlab.clone_with") == "ssh" {
		href = strings.ReplaceAll(
			viper.GetString("integrations.gitlab.ssh_format"),
			"[.RepoFullName]",
			pushEvent.Project.PathWithNamespace,
		)
	}

	version := strings.ReplaceAll(
		pushEvent.Ref,
		"refs/tags/",
		"",
	)

	releaseRequest := model.ReleaseRequest{
		Name:    pushEvent.Project.Name,
		URL:     href,
		Version: version,
	}

	passToWorker(
		c,
		messages,
		logger,
		releaseRequest,
	)
}