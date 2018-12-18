// You can customize this file, wont be replace
type AllRepositories struct {
    {{- range $k, $v := reponames }}
    repositories.I{{ $v }}
    {{- end}}
}

func sampleInject(db *sqlx.DB) AllRepositories {
    wire.Build(
        {{- range $k, $v := reponames }}
        repositories.New{{ $v }},
        {{- end}}
        AllRepositories{},
    )
    return AllRepositories{}
}