package repository

import (
	"bytes"
	"fmt"
	"io/ioutil"
	"strings"
	"testing"
	"time"

	"github.com/google/go-github/v39/github"
	"github.com/jonboulle/clockwork"
	"github.com/openshift-pipelines/pipelines-as-code/pkg/apis/pipelinesascode/v1alpha1"
	"github.com/openshift-pipelines/pipelines-as-code/pkg/cli"
	"github.com/openshift-pipelines/pipelines-as-code/pkg/params"
	"github.com/openshift-pipelines/pipelines-as-code/pkg/params/clients"
	"github.com/openshift-pipelines/pipelines-as-code/pkg/params/info"
	testclient "github.com/openshift-pipelines/pipelines-as-code/pkg/test/clients"
	"gotest.tools/v3/golden"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	knativeapis "knative.dev/pkg/apis"
	"knative.dev/pkg/apis/duck/v1beta1"
	rtesting "knative.dev/pkg/reconciler/testing"
)

func newIOStream() (*cli.IOStreams, *bytes.Buffer) {
	in := &bytes.Buffer{}
	out := &bytes.Buffer{}
	errOut := &bytes.Buffer{}
	return &cli.IOStreams{
		In:     ioutil.NopCloser(in),
		Out:    out,
		ErrOut: errOut,
	}, out
}

func TestDescribe(t *testing.T) {
	cw := clockwork.NewFakeClock()
	type args struct {
		currentNamespace string
		repoName         string
		statuses         []v1alpha1.RepositoryRunStatus
		opts             *cli.PacCliOpts
	}
	tests := []struct {
		name    string
		args    args
		wantErr bool
	}{
		{
			name: "Describe a Pipeline with a Single Run",
			args: args{
				repoName:         "test-run",
				currentNamespace: "namespace",
				opts:             &cli.PacCliOpts{},
				statuses: []v1alpha1.RepositoryRunStatus{
					{
						Status: v1beta1.Status{
							Conditions: []knativeapis.Condition{
								{
									Reason: "Success",
								},
							},
						},
						PipelineRunName: "pipelinerun1",
						StartTime:       &metav1.Time{Time: cw.Now().Add(-16 * time.Minute)},
						CompletionTime:  &metav1.Time{Time: cw.Now().Add(-15 * time.Minute)},
						SHA:             github.String("SHA"),
						SHAURL:          github.String("https://anurl.com/commit/SHA"),
						Title:           github.String("A title"),
						TargetBranch:    github.String("TargetBranch"),
					},
				},
			},
			wantErr: false,
		},
		{
			name: "Describe a Pipeline with a Single Run - optnamespace",
			args: args{
				repoName:         "test-run",
				currentNamespace: "namespace",
				opts: &cli.PacCliOpts{
					Namespace: "optnamespace",
				},
				statuses: []v1alpha1.RepositoryRunStatus{
					{
						Status: v1beta1.Status{
							Conditions: []knativeapis.Condition{
								{
									Reason: "Success",
								},
							},
						},
						PipelineRunName: "pipelinerun1",
						StartTime:       &metav1.Time{Time: cw.Now().Add(-16 * time.Minute)},
						CompletionTime:  &metav1.Time{Time: cw.Now().Add(-15 * time.Minute)},
						SHA:             github.String("SHA"),
						SHAURL:          github.String("https://anurl.com/commit/SHA"),
						Title:           github.String("A title"),
						TargetBranch:    github.String("TargetBranch"),
					},
				},
			},
			wantErr: false,
		},
		{
			name: "Describe a Pipeline with a Multiple Run",
			args: args{
				opts:             &cli.PacCliOpts{},
				repoName:         "test-run",
				currentNamespace: "namespace",
				statuses: []v1alpha1.RepositoryRunStatus{
					{
						Status: v1beta1.Status{
							Conditions: []knativeapis.Condition{
								{
									Reason: "Success",
								},
							},
						},
						PipelineRunName: "pipelinerun1",
						StartTime:       &metav1.Time{Time: cw.Now().Add(-16 * time.Minute)},
						CompletionTime:  &metav1.Time{Time: cw.Now().Add(-15 * time.Minute)},
						SHA:             github.String("SHA"),
						SHAURL:          github.String("https://anurl.com/commit/SHA"),
						Title:           github.String("A title"),
						TargetBranch:    github.String("TargetBranch"),
						EventType:       github.String("pull_request"),
					},
					{
						Status: v1beta1.Status{
							Conditions: []knativeapis.Condition{
								{
									Reason: "Success",
								},
							},
						},
						PipelineRunName: "pipelinerun2",
						StartTime:       &metav1.Time{Time: cw.Now().Add(-18 * time.Minute)},
						CompletionTime:  &metav1.Time{Time: cw.Now().Add(-17 * time.Minute)},
						SHA:             github.String("SHA2"),
						SHAURL:          github.String("https://anurl.com/commit/SHA2"),
						Title:           github.String("Another Update"),
						TargetBranch:    github.String("TargetBranch"),
						EventType:       github.String("pull_request"),
					},
				},
			},
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ns := tt.args.currentNamespace
			if tt.args.opts.Namespace != "" {
				ns = tt.args.opts.Namespace
			}
			repositories := []*v1alpha1.Repository{
				{
					ObjectMeta: metav1.ObjectMeta{
						Name:      tt.args.repoName,
						Namespace: ns,
					},
					Spec: v1alpha1.RepositorySpec{
						URL: "https://anurl.com",
					},
					Status: tt.args.statuses,
				},
			}

			tdata := testclient.Data{
				Namespaces: []*corev1.Namespace{
					{
						ObjectMeta: metav1.ObjectMeta{
							Name: tt.args.currentNamespace,
						},
					},
				},
				Repositories: repositories,
			}
			ctx, _ := rtesting.SetupFakeContext(t)
			stdata, _ := testclient.SeedTestData(t, ctx, tdata)
			cs := &params.Run{
				Clients: clients.Clients{
					PipelineAsCode: stdata.PipelineAsCode,
				},
				Info: info.Info{Kube: info.KubeOpts{Namespace: tt.args.currentNamespace}},
			}

			io, out := newIOStream()
			if err := describe(
				ctx, cs, cw, tt.args.opts, io,
				tt.args.repoName); (err != nil) != tt.wantErr {
				t.Errorf("describe() error = %v, wantErr %v", err, tt.wantErr)
			} else {
				golden.Assert(t, out.String(), strings.ReplaceAll(fmt.Sprintf("%s.golden", t.Name()), "/", "-"))
			}
		})
	}
}