package archive

import (
	backupv1alpha1 "git.vshn.net/vshn/baas/apis/backup/v1alpha1"
	"git.vshn.net/vshn/baas/service"
	batchv1 "k8s.io/api/batch/v1"
	corev1 "k8s.io/api/core/v1"
)

type byCreationTime []backupv1alpha1.Archive

func (b byCreationTime) Len() int      { return len(b) }
func (b byCreationTime) Swap(i, j int) { b[i], b[j] = b[j], b[i] }

func (b byCreationTime) Less(i, j int) bool {

	if b[i].CreationTimestamp.Equal(&b[j].CreationTimestamp) {
		return b[i].Name < b[j].Name
	}

	return b[i].CreationTimestamp.Before(&b[j].CreationTimestamp)
}

func newArchiveJob(archive *backupv1alpha1.Archive, config config) *batchv1.Job {

	args := []string{"-archive", "-restoreType", "s3"}

	job := service.GetBasicJob("archive", config.GlobalConfig, &archive.ObjectMeta)
	job.Spec.Template.Spec.Containers[0].Args = args
	finalEnv := append(job.Spec.Template.Spec.Containers[0].Env, setUpEnvVariables(archive, config)...)
	job.Spec.Template.Spec.Containers[0].Env = finalEnv

	return job
}

func setUpEnvVariables(archive *backupv1alpha1.Archive, config config) []corev1.EnvVar {
	vars := []corev1.EnvVar{}

	vars = append(vars, service.BuildS3EnvVars(archive.GlobalOverrides.RegisteredBackend.S3, config.GlobalConfig)...)

	vars = append(vars, service.BuildRepoPasswordVar(archive.GlobalOverrides.RegisteredBackend.RepoPasswordSecretRef, config.GlobalConfig))

	if archive.Spec.RestoreMethod.S3 != nil {
		vars = append(vars, service.BuildRestoreS3Env(archive.Spec.RestoreMethod.S3, config.GlobalConfig)...)
	}

	return vars
}
