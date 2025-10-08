package cloudfront

import (
	"fmt"

	"github.com/aws/aws-cdk-go/awscdk/v2"
	"github.com/aws/aws-cdk-go/awscdk/v2/awscloudfront"
	"github.com/aws/aws-cdk-go/awscdk/v2/awss3"
	"github.com/aws/constructs-go/constructs/v10"
)

// -----------------------------------------------------------------------------
// ENUM: Tipos de origen soportados
// -----------------------------------------------------------------------------
type OriginType string

const (
	OriginTypeS3  OriginType = "S3"
	OriginTypeAPI OriginType = "API"
	OriginTypeALB OriginType = "ALB"
)

// -----------------------------------------------------------------------------
// Propiedades principales para crear una distribución CloudFront
// -----------------------------------------------------------------------------
type CloudFrontPropertiesV2 struct {
	OriginType OriginType

	// Recursos posibles (solo uno debe estar presente según el tipo)
	S3Bucket     awss3.IBucket
	ApiOrigin    awscloudfront.IOrigin
	LoadBalancer awscdk.Resource

	// Configuración opcional
	DomainNames                 []string
	CertificateArn              string
	WebAclArn                   string
	Comment                     string
	EnableAccessLogging         bool
	AutoConfigureS3BucketPolicy bool
}

// -----------------------------------------------------------------------------
// CloudFrontFactory — punto de entrada para crear distribuciones CloudFront
// -----------------------------------------------------------------------------
func NewDistributionV2(scope constructs.Construct, id string, props CloudFrontPropertiesV2) awscloudfront.Distribution {
	var strategy CloudFrontStrategy

	// Selecciona el Strategy según el tipo de origen
	switch props.OriginType {
	case OriginTypeS3:
		if props.S3Bucket == nil {
			panic("Debe proporcionar S3Bucket para una distribución S3")
		}
		strategy = &S3CloudFrontStrategy{}

	case OriginTypeAPI:
		if props.ApiOrigin == nil {
			panic("Debe proporcionar ApiOrigin para una distribución API")
		}
		//strategy = &APICloudFrontStrategy{}
		panic("API strategy no implementada aún")
	case OriginTypeALB:
		if props.LoadBalancer == nil {
			panic("Debe proporcionar LoadBalancer para una distribución ALB")
		}
		//strategy = &ALBCloudFrontStrategy{}
		panic("ALB strategy no implementada aún")
	default:
		panic(fmt.Sprintf("Origen no soportado: %s", props.OriginType))
	}

	// Construye y devuelve la distribución usando el strategy seleccionado
	return strategy.Build(scope, id, props)
}
