#!/bin/bash

# =============================================================================
# AWS CDK Static Website Deployment Script (Go/No-Hardcoding)
# =============================================================================
# Este script automatiza el proceso completo de despliegue,
# usando SIEMPRE las credenciales y configuración por defecto (AWS_PROFILE/REGION)
# para evitar valores hardcodeados.
# =============================================================================

set -e

echo "🚀 Iniciando despliegue completo del sitio web estático (Go CDK)..."

# Colores para output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

# Configuración del proyecto
DIST_DIR="./stacks/website/dist"
CDK_STACK_NAME="DevStaticWebsite"
AWS_PROFILE="default"

# ===============================
# Helper Functions
# ===============================
log() { echo -e "${GREEN}[INFO]${NC} $1"; }
warn() { echo -e "${YELLOW}[WARN]${NC} $1"; }
error() { echo -e "${RED}[ERROR]${NC} $1"; exit 1; }
debug() { echo -e "${BLUE}[DEBUG]${NC} $1"; }
print_header() { echo -e "\n${BLUE}================================${NC}"; echo -e "${BLUE}$1${NC}"; echo -e "${BLUE}================================${NC}\n"; }

# ===============================
# CONFIGURACIÓN AUTOMÁTICA DE CDK (No-Hardcoding)
# ===============================
print_header "1. Configurando variables de Entorno"

# Obtener Account ID automáticamente
CDK_DEFAULT_ACCOUNT=$(aws sts get-caller-identity --query Account --output text --profile $AWS_PROFILE 2>/dev/null)
if [ -z "$CDK_DEFAULT_ACCOUNT" ]; then
    error "No se pudo obtener el Account ID de AWS CLI. Verifica tu perfil '$AWS_PROFILE'."
fi

# Obtener región automáticamente desde la configuración de la CLI
CDK_DEFAULT_REGION=$(aws configure get region --profile $AWS_PROFILE)
if [ -z "$CDK_DEFAULT_REGION" ]; then
    warn "No se pudo obtener la región por defecto de AWS CLI. Usando 'us-east-1' como fallback."
    CDK_DEFAULT_REGION="us-east-1"
fi

# Exportar las variables para que el código Go y la CLI las usen
export CDK_DEFAULT_ACCOUNT
export CDK_DEFAULT_REGION

log "✅ Configuración CDK obtenida automáticamente:"
log "    Account (CDK_DEFAULT_ACCOUNT): $CDK_DEFAULT_ACCOUNT"
log "    Region (CDK_DEFAULT_REGION): $CDK_DEFAULT_REGION"

# ===============================
# Chequeos de Herramientas
# ===============================
print_header "2. Chequeos de Herramientas"

if ! command -v go &> /dev/null; then error "Go no está instalado. Instálalo primero."; fi
log "Go instalado: $(go version)"

if ! command -v cdk &> /dev/null; then error "CDK CLI no está instalado. Instala con: npm install -g aws-cdk"; fi
log "CDK instalado: $(cdk --version)"

# =============================================================================
# CDK Bootstrap (Verificación y Ejecución) - LÓGICA MODIFICADA
# =============================================================================
print_header "3. Verificación de CDK Bootstrap"

log "Comprobando si el entorno está bootstrapped en la región $CDK_DEFAULT_REGION..."

# Verificar si el stack de Bootstrap existe y está completo.
BOOTSTRAP_STATUS=$(aws cloudformation describe-stacks \
    --stack-name CDKToolkit \
    --region $CDK_DEFAULT_REGION \
    --profile $AWS_PROFILE \
    --query 'Stacks[0].StackStatus' \
    --output text 2>/dev/null)

# Revisamos si el stack NO está en un estado COMPLETO.
if [ "$BOOTSTRAP_STATUS" != "CREATE_COMPLETE" ] && [ "$BOOTSTRAP_STATUS" != "UPDATE_COMPLETE" ]; then
    warn "CDK no está completamente bootstrapped o no existe. Ejecutando bootstrap ahora..."
    
    # Ejecuta el bootstrap. El comando retorna inmediatamente, iniciando la creación en CFN.
    cdk bootstrap \
        --profile $AWS_PROFILE \
        --cloudformation-execution-policies arn:aws:iam::aws:policy/AdministratorAccess 
    
    if [ $? -ne 0 ]; then
        error "Falló la ejecución de 'cdk bootstrap'."
    fi

    # **CLAVE:** Esperar a que CloudFormation termine la creación del stack CDKToolkit.
    log "Esperando a que el stack CDKToolkit se complete..."
    aws cloudformation wait stack-create-complete \
        --stack-name CDKToolkit \
        --region $CDK_DEFAULT_REGION \
        --profile $AWS_PROFILE

    if [ $? -ne 0 ]; then
        error "Falló la espera del stack CDKToolkit. CloudFormation no pudo crear el stack."
    fi
fi

log "✅ CDK está bootstrapped y listo para la publicación de activos."

# ===============================
# Preparación del Contenido
# ===============================
print_header "4. Preparación de Contenido y Dependencias"

# Preparación de directorios (usando la lógica de tu script original)
if [ ! -d "$DIST_DIR" ]; then mkdir -p "$DIST_DIR"; log "Creando directorio: $DIST_DIR"; fi
mkdir -p "$DIST_DIR/assets/css"
mkdir -p "$DIST_DIR/assets/js"
mkdir -p "$DIST_DIR/assets/images"
log "Directorios de assets creados/verificados."

# Chequear index.html
if [ ! -f "$DIST_DIR/index.html" ]; then
    error "index.html no encontrado. Por favor, crea el archivo en $DIST_DIR/"
fi
log "Encontrado index.html listo para el despliegue."

# Go dependencies
log "Construyendo dependencias de Go..."
go mod tidy
go mod download
log "✅ Dependencias Go actualizadas."

# ===============================
# Despliegue de la Infraestructura
# ===============================
print_header "5. Despliegue de CDK a AWS"

log "Sintetizando stack $CDK_STACK_NAME..."
cdk synth $CDK_STACK_NAME --profile $AWS_PROFILE

log "Desplegando stack $CDK_STACK_NAME. Espera la confirmación..."
warn "Presiona Ctrl+C para cancelar, o espera 5 segundos para continuar..."
sleep 5

# Deploy con las variables de entorno consistentes
cdk deploy $CDK_STACK_NAME \
    --profile $AWS_PROFILE \
    --require-approval never

if [ $? -ne 0 ]; then
    error "Falló el despliegue del stack $CDK_STACK_NAME"
fi

log "✅ Stack $CDK_STACK_NAME desplegado exitosamente"

# ===============================
# Obtener Outputs y Diagnóstico de S3
# ===============================
print_header "6. Stack Outputs y Diagnóstico de S3"

# Esperar un poco para que CloudFormation se estabilice
sleep 10

# Obtener URL del sitio web (asumiendo que tu stack de Go exporta WebsiteURL)
WEBSITE_URL=$(aws cloudformation describe-stacks \
    --stack-name dev-static-website \
    --query 'Stacks[0].Outputs[?OutputKey==`WebsiteURL`].OutputValue' \
    --output text \
    --region $CDK_DEFAULT_REGION)

DISTRIBUTION_ID=$(aws cloudformation describe-stacks \
    --stack-name dev-static-website \
    --query 'Stacks[0].Outputs[?OutputKey==`DistributionId`].OutputValue' \
    --output text \
    --region $CDK_DEFAULT_REGION)
    
BUCKET_NAME=$(aws cloudformation describe-stacks \
    --stack-name dev-static-website \
    --query 'Stacks[0].Outputs[?OutputKey==`BucketName`].OutputValue' \
    --output text \
    --region $CDK_DEFAULT_REGION)

if [ -n "$BUCKET_NAME" ]; then
    log "Verificando contenido en S3 bucket: $BUCKET_NAME"
    
    # Contar objetos en el bucket para diagnóstico
    OBJECT_COUNT=$(aws s3 ls s3://$BUCKET_NAME --recursive --profile $AWS_PROFILE 2>/dev/null | wc -l)
    
    if [ "$OBJECT_COUNT" -lt 1 ]; then
        warn "⚠️ ATENCIÓN: El bucket S3 está vacío ($OBJECT_COUNT objetos)."
        warn "Esto es la causa más probable del error 'NotFound'."
        warn "Acciones: 1. Verifica la ruta 'SourcePath' en tu código Go. 2. Vuelve a ejecutar el despliegue."
    else
        log "✅ El bucket S3 contiene $OBJECT_COUNT archivos."
        log "Si el error 'NotFound' persiste, revisa si CloudFront tiene 'index.html' como Default Root Object."
        
        log "Listado de primeros 5 archivos:"
        aws s3 ls s3://$BUCKET_NAME --recursive --profile $AWS_PROFILE | head -n 5
    fi
fi

if [ -n "$WEBSITE_URL" ]; then
    echo ""
    echo "🎉 ¡Despliegue completado exitosamente!"
    echo "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"
    echo "🌐 Tu sitio web (CloudFront) está disponible en: $WEBSITE_URL"
    echo "⏳ Nota: CloudFront puede tardar 5-10 minutos en propagar"
    echo "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"
    echo ""
else
    warn "No se pudo obtener la URL del sitio web."
fi

log "✅ Script de despliegue finalizado."