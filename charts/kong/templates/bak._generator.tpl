{{- define "ingress-routes-map" }}
  {{- $REQ_SVC_CLASS_MAP := dict }}
  {{- range $yamlFilePath, $_ :=  .Files.Glob  "routes/*.yaml" }}
    {{- $SERVICE_NAME := (trimPrefix "routes/" $yamlFilePath | trimSuffix ".yaml")}}
    {{- range $routePath, $routeProps := ($.Files.Get $yamlFilePath | fromYaml).apis }}

      {{- $REQUEST_METHODS := $routeProps.request_method | join "-"}}

      {{- $REQ_CLASS := "AUTH"}}
      {{- if hasKey $routeProps "auth" }}
        {{- $REQ_CLASS = "NOAUTH"}}
      {{- else if hasKey $routeProps "staff" }}
        {{- $REQ_CLASS = "STAFF"}}
      {{- end}}

      {{- if not (hasKey $REQ_SVC_CLASS_MAP $REQUEST_METHODS)}}
        {{- $_ := set $REQ_SVC_CLASS_MAP $REQUEST_METHODS dict}}
      {{- end}}

      {{- $tempClassMap := (get $REQ_SVC_CLASS_MAP $REQUEST_METHODS)}}
      {{- if not (hasKey $tempClassMap $REQ_CLASS)}}
        {{- $_ := set $tempClassMap $REQ_CLASS dict}}
      {{- end}}
      
      {{- $tempSRVMap := (get $tempClassMap $REQ_CLASS)}}
      {{- if not (hasKey $tempSRVMap $SERVICE_NAME)}}
        {{- $_ := set $tempSRVMap $SERVICE_NAME $routePath}}
      {{- else}}
        {{- $_ := set $tempSRVMap $SERVICE_NAME (printf "%s|%s" (get $tempSRVMap $SERVICE_NAME) $routePath)}}
      {{- end}}

      {{- $_ := set $tempClassMap $REQ_CLASS $tempSRVMap}}
      {{- $_ := set $REQ_SVC_CLASS_MAP $REQUEST_METHODS $tempClassMap}}

    {{- end}}
  {{- end}}

  {{$REQ_SVC_CLASS_MAP | toYaml}}
{{- end}}


{{- define "api-ingress-manifests" }}
  {{- $REQ_SVC_CLASS_MAP := (include "ingress-routes-map" . | trim | fromYaml)}}
  {{- range $reqMethods, $reqClassMap := $REQ_SVC_CLASS_MAP}}
    {{- range $reqClass, $reqSrvMap := $reqClassMap}}
apiVersion: networking.k8s.io/v1
kind: Ingress
metadata:
  namespace: {{$.Values.namespace}}
  name: {{$reqClass | lower}}-{{$reqMethods | lower}}-ingress
  annotations:
    kubernetes.io/ingress.class: {{$.Values.namespace}}
    konghq.com/override: {{$reqClass | lower}}-{{$reqMethods | lower}}-ingressOverride
    {{- $scheme := "https"}}
    {{- if (eq $reqClass "STAFF")}}
      {{- $scheme = "http"}}
    {{- else if (eq $reqClass "AUTH")}}
    konghq.com/plugins: app-jwt-ses
    {{- end}}
spec:
  tls:
    - hosts:
    {{- range $srvName := (keys $reqSrvMap)}}
      - {{$.Values.apiPrefix}}-{{get (get $.Values.apiConfig $srvName) "apiSubDomain"}}.sentieo.com
    {{- end}}
  rules:
    - http:
        paths:
          {{- range $srvName, $reqPaths := $reqSrvMap}}
            {{- $formatString := "/api/(%s)"}}
            {{- if (not (contains "|" $reqPaths))}}
              {{- $formatString = "/api/%s"}}
            {{- end}}
          - path: {{printf $formatString $reqPaths}}
            pathType: Prefix
            backend:
              service:
                name: {{$srvName}}
                port:
                  number: {{get (get $.Values.apiConfig $srvName) "containerPort"}}
          {{- end}}
---
apiVersion: configuration.konghq.com/v1
kind: KongIngress
metadata:
  name: {{$reqClass | lower}}-{{$reqMethods | lower}}-ingressOverride
  namespace: {{$.Values.namespace}}
proxy:
  protocol: {{$scheme}}
  connect_timeout: 10000
  retries: 10
  read_timeout: 10000
  write_timeout: 10000
route:
  methods:
  {{- range $method := (split "-" $reqMethods)}}
    - {{$method}}
  {{- end}}
  regex_priority: 0
  strip_path: false
  preserve_host: true
  protocols:
  - {{$scheme}}
  {{- if (eq $scheme "https")}}
  https_redirect_status_code: 302
  {{- end}}
---
    {{- end}}
  {{- end}}
{{- end}}
