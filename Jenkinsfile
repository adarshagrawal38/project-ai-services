pipeline {
    agent { label 'ai-services-4-spyre-card-1' }

    environment {
        PROJECT_NAME = 'project-ai-services'
        AI_SERVICES_DIR = 'ai-services'
        AI_SERVICES_BINARY = './bin/ai-services'
        MAX_APP_RUN_TIME_IN_MINS = '1'
    }

    parameters {
        string(
            name: 'CHECKOUT',
            description: 'Pull request number or commit hash to build ai-services'
        )
        booleanParam(
            name: 'BUILD_RAG_BACKEND',
            description: 'Build rag backend image'
        )
        booleanParam(
            name: 'BUILD_RAG_UI',
            description: 'Build rag UI image'
        )
        choice(
            name: 'APP',
            choices: ['rag-dev', 'summarization'],
            description: 'Select desired application which you want to deploy'
        )
        stashedFile 'INGEST_DOC_FILE'
    }

    // Using options to allow one deployment at any given point of time.
    options {
        disableConcurrentBuilds()
    }

    stages {
        stage('Validate parameters') {
            steps {
                script {
                    if (!params.CHECKOUT.trim()) {
                        error('PR number or commit hash must be provided')
                    }
                }
            }
        }

        stage('Checkout Code') {
            steps {
                script {
                    deleteDir()
                    sh "pwd"
                    repoCheckout(params.CHECKOUT)
                }
            }
        }

        

        stage('Build AI Services Binary') {
            steps {
                sh '''
                    cd ${PROJECT_NAME}/${AI_SERVICES_DIR}
                    make build
                    ${AI_SERVICES_BINARY} --version
                '''
            }
        }

        // Check if there is an existing application
        stage('Delete App') {
            steps{
                script {
                    checkMachineStatus()
                    cleanupMachine()
                }
            }
        }

        stage('Build Selected Image') {
            steps {
                script {
                    buildAppImage(params.APP)
                }
            }
        }

        stage('Deployment') {
            steps {
                sh '''
                    cd ${PROJECT_NAME}/${AI_SERVICES_DIR}
                    ${AI_SERVICES_BINARY} application create ${APP}-cicd -t ${APP}
                '''
            }
        }
        stage('Ingest DOC') {
            steps {
                script {
                
                unstash 'INGEST_DOC_FILE'
                    
                sh '''
                    mv INGEST_DOC_FILE /var/lib/ai-services/applications/${APP}-cicd/docs/doc.pdf
                    echo "ingest DOC"
                    cd ${PROJECT_NAME}/${AI_SERVICES_DIR}
                    ${AI_SERVICES_BINARY} application start rag-test --pod=${APP}-cicd--ingest-docs -y
                '''
                }
                
            }
        }

        stage('Test your PR') {
            steps{
                script{
                    // Polling to check if app is deleted
                    for (int i = 0; i < env.MAX_APP_RUN_TIME_IN_MINS.toInteger() ; i++) {
                        def appName = getRunningApp()
                        if (appName.isEmpty()) {
                            break
                        }
                        println "Iteration number ${i}, waiting for 60s"
                        sh 'sleep 60s'
                    }
                }
            }
        }

    }
}

def getRunningApp() {
    def appName = ""
    dir("${env.PROJECT_NAME}/${env.AI_SERVICES_DIR}") {
        def apps = sh(
            script: "./bin/ai-services application ps 2>&1",
            returnStdout: true
        ).trim()
        def outputLines = apps.readLines()
        if (outputLines.size() > 2 ) {
            appName = outputLines[2].split()[0]
            echo "${appName}"
        }
    }
    return appName
}

def repoCheckout(String branch) {
    sh "git clone https://github.com/IBM/project-ai-services.git"
    dir("${PROJECT_NAME}") {
        if (branch ==~ /^\d+$/) {
            println "Checking out to ${branch} PR number"
            sh """
                git fetch origin pull/${branch}/head:pr-${branch}
                git checkout pr-${branch}
            """
        }else {
            println "Checking out to ${branch} commit hash"
            sh "git rev-parse --verify ${branch}"
            sh "git checkout ${branch}"
        }
    }
}

// Check if machine is free or not for deployment of application
def checkMachineStatus() {
    def appName = getRunningApp()
    dir("${env.PROJECT_NAME}/${env.AI_SERVICES_DIR}") {
        if (appName.isEmpty() ) {
            println "No applications running in the machine."
            println "Machine is in ready state for deployment."
        }else {
            if (!appName.contains("cicd")) {
                println "Please delete ${appName} to proceed for deployment"
                error("Please delete ${appName} to proceed for deployment")
            }
        }
    }
}

def cleanupMachine() {
    def appName = getRunningApp()
    if (appName.isEmpty()) {
        return
    }
    dir("${env.PROJECT_NAME}/${env.AI_SERVICES_DIR}") {
        if (appName.contains("cicd")) {
            println "Cleaning up ${appName} which is deployed by pipelines."
            sh "./bin/ai-services application delete ${appName} -y"
        }else{
            println "Please delete ${appName} from the machine."
            error("Please delete ${appName} to proceed for deployment")
        }
    }
}

def buildAppImage(String appName) {
    if (appName == "rag-dev") {
        if(params.BUILD_RAG_BACKEND) {
            def imageVal = buildImage(params.APP, "spyre-rag/src")
            updateYamlFile(params.APP, imageVal, ".backend.image")
        }
        if(params.BUILD_RAG_UI) {
            def imageVal = buildImage("rag-ui", "spyre-rag/ui")
            updateYamlFile(params.APP, imageVal, ".ui.image")
        }
    }else {
        error("Selected ${appName} not supported.")
    }
    
}

def buildImage(String imageName, String containerFilePath) {
    String localRegistry = 'localhost'
    String imagePath = ""
    dir(env.PROJECT_NAME) {
        sh "git rev-parse --short HEAD"
        dir(containerFilePath) {
            // Build image
            sh "make build REGISTRY=${localRegistry}"

            // Get the tag from Makefile
            def tag = sh(script: "make image-tag REGISTRY=${localRegistry}", returnStdout: true).trim()
            imagePath = "${localRegistry}/${imageName}:${tag}"
            echo "Backend image set to ${imagePath}"
        }
    }
    return imagePath
}

// Method to update local image in the yaml file
// so that deployment of local image is done
def updateYamlFile(String appName,String imageValue, String overridePath) {
    dir(env.PROJECT_NAME) {
        def valuesFile = "ai-services/assets/applications/${appName}/values.yaml"
        if (!fileExists(valuesFile)) {
            error("no values.yaml for ${appName}")
            return
        }
        sh "yq e '${overridePath} = \"${imageValue}\"' -i ${valuesFile}"
        echo "Updated image values"
    }
}