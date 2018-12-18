// Sample code, dont use
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