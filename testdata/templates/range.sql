SELECT * FROM `user`
{{range $key, $value := .}}
{{$key}} : {{$value}}
{{end}}