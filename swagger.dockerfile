FROM swaggerapi/swagger-ui
COPY ./ConnectorEdgeCloud.yaml /foo/swagger.yaml
ENV SWAGGER_JSON=/foo/swagger.yaml
