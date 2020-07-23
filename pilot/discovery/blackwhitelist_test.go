package discovery

import (
	"testing"
)

func TestBlackWhiteList_IsResponsible(t *testing.T) {
	type fields struct {
		bListNS  string
		wListNS  string
		bListPod string
		wListPod string
	}
	type args struct {
		namespace string
		pod       string
	}

	const (
		defaultNS        = "default"
		kubeNS           = "kube-system"
		defaultAndKubeNS = "default,kube-system"
		PodListRegex     = "^kube-system/(lb-.+|apiserver-proxy-nginx-preset-.+)$"
		lbPodRegex       = "^kube-system/(lb-3160836842-proxy-nginx-eqygq-cfb79fccf-s5zz5.*)$"
		apiServerPodName = "apiserver-proxy-nginx-preset-7db45cfddb-bsk97"
		lbPodName        = "lb-3160836842-proxy-nginx-eqygq-cfb79fccf-s5zz5"
	)

	tests := []struct {
		name    string
		fields  fields
		args    args
		want    bool
		wantErr bool
	}{
		{
			name: "priority in blacklist and whitelist pod",
			fields: fields{
				bListNS:  kubeNS,
				wListNS:  defaultNS,
				bListPod: lbPodRegex,
				wListPod: PodListRegex,
			},
			args: args{
				namespace: kubeNS,
				pod:       lbPodName,
			},
			want: false,
		},
		{
			name: "no pod name",
			fields: fields{
				bListNS:  kubeNS,
				wListPod: PodListRegex,
			},
			args: args{
				namespace: kubeNS,
			},
			want: false,
		},
		{
			name: "priority in whitelist pod and blacklist namespace\"",
			fields: fields{
				bListNS:  defaultAndKubeNS,
				wListPod: PodListRegex,
			},
			args: args{
				namespace: kubeNS,
				pod:       apiServerPodName,
			},
			want: true,
		},

		{
			name: "priority in blacklist pod and whitelist namespace",
			fields: fields{
				bListNS:  defaultNS,
				wListNS:  kubeNS,
				bListPod: PodListRegex,
			},
			args: args{
				namespace: kubeNS,
				pod:       apiServerPodName,
			},
			want: false,
		},
		{
			name: "in the whitelist, not in the blacklist namespace",
			fields: fields{
				bListNS: defaultNS,
				wListNS: defaultAndKubeNS,
			},
			args: args{
				namespace: kubeNS,
				pod:       lbPodName,
			},
			want: true,
		},
		{
			name: "in the whitelist, not in the blacklist namespace",
			fields: fields{
				bListNS: defaultNS,
				wListNS: defaultAndKubeNS,
			},
			args: args{
				namespace: defaultNS,
				pod:       lbPodName,
			},
			want: false,
		},
		{
			name: "in the blacklist, not in the whitelist namespace",
			fields: fields{
				bListNS: kubeNS,
				wListNS: defaultNS,
			},
			args: args{
				namespace: kubeNS,
				pod:       lbPodName,
			},
			want: false,
		},
		{
			name: "both in the blacklist and the whitelist namespace",
			fields: fields{
				bListNS: kubeNS,
				wListNS: defaultAndKubeNS,
			},
			args: args{
				namespace: kubeNS,
				pod:       lbPodName,
			},
			want: false,
		},
		{
			name: "empty whitelist namespace",
			fields: fields{
				bListNS: defaultNS,
			},
			args: args{
				namespace: kubeNS,
				pod:       lbPodName,
			},
			want: true,
		},
		{
			name: "empty blacklist and whitelist namespace",
			args: args{
				namespace: kubeNS,
				pod:       lbPodName,
			},
			want: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			b := &BlackWhiteList{
				bListNS:  tt.fields.bListNS,
				wListNS:  tt.fields.wListNS,
				bListPod: tt.fields.bListPod,
				wListPod: tt.fields.wListPod,
			}
			got, err := b.IsResponsible(tt.args.namespace, tt.args.pod)
			if err != nil {
				t.Errorf("IsResponsible() error = %v", err)
				return
			}
			if got != tt.want {
				t.Errorf("IsResponsible() got = %v, want %v", got, tt.want)
			}
		})
	}
}
