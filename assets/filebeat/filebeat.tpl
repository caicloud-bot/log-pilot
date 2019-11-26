{{range .configList}}
- type: log
  enabled: true
  paths:
  - {{ .LogFile }}
  scan_frequency: 10s
  fields_under_root: true
  {{if .Stdout}}
  docker-json:
    stream: all
    partial: true 
    cri_flags: true
  {{end}}
  fields:
    cluster: ${CLUSTER_ID}
    {{- range $key, $value := .Tags }}
    {{ $key }}: "{{ $value }}"
    {{- end }}
  tail_files: false
  # Harvester closing options
  close_eof: false
  close_inactive: 5m
  close_removed: false
  close_renamed: false
  ignore_older: 48h  
  # State options
  clean_removed: true
  clean_inactive: 72h
  # If an ES receiver failed to index a message, Filebeat will output an error message matching the format below;
  # this error message could be gathered again and thus form a loop, and likely consume a lot of resources;
  # therefore, we exclude this kind of error message when gathering logs from a Filebeat container
  {{if $.isFilebeat}}
  exclude_lines: ['^[0-9]{4}-[0-9]{1,2}-[0-9]{1,2}T.+WARN.+elasticsearch/client.+Cannot index event publisher.+']
  {{end}}
{{- end}}

