# The next line updates PATH for the Google Cloud SDK.
if [ -f /usr/local/share/google/google-cloud-sdk/path.bash.inc ]; then
source '/usr/local/share/google/google-cloud-sdk/path.bash.inc'
fi

if [ -f /usr/local/share/google/google-cloud-sdk/completion.bash.inc ];then
# The next line enables shell command completion for gcloud.
source '/usr/local/share/google/google-cloud-sdk/completion.bash.inc'
else
source '/usr/lib64/google-cloud-sdk/completion.bash.inc'
fi
