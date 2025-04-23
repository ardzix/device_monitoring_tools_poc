pipeline {
    agent { label 'escrow prod' }

    environment {
        DEPLOY = 'true'
        DOCKER_IMAGE = 'ardzix/employee_monitoring_host' // DockerHub repo
        DOCKER_REGISTRY_CREDENTIALS = 'ard-dockerhub' // Jenkins credentials ID for DockerHub PAT
        NAMESPACE = 'employee_monitoring_host' // Service name
        STACK_NAME = 'employee_monitoring_host' // Swarm stack name
        REPLICAS = '2' // Number of service replicas
        NETWORK_NAME = 'production' // Swarm overlay network
    }

    stages {
        stage('Clean Workspace') {
            steps {
                script {
                    sh 'rm -rf ./* ./.??* || true'
                    sh 'ls -la'
                }
            }
        }

        stage('Checkout Code') {
            steps {
                script {
                    sh 'git clone https://github.com/ardzix/device_monitoring_tools_poc.git .'
                }
            }
        }

        stage('Inject Environment Variables') {
            steps {
                script {
                    withCredentials([
                        file(credentialsId: 'employee-monitoring-host-env', variable: 'ENV_FILE'),
                        string(credentialsId: 'ms-arnatech-storage-access', variable: 'AWS_ACCESS_KEY_ID'),
                        string(credentialsId: 'ms-arnatech-storage-secret', variable: 'AWS_SECRET_ACCESS_KEY')
                    ]) {
                        sh '''
                            # Create monitoring-host directory if it doesn't exist
                            mkdir -p ./monitoring-host
                            
                            # Copy and update the .env file
                            cp $ENV_FILE ./monitoring-host/.env
                            
                            # Update S3 credentials in .env file
                            sed -i "s|^AWS_ACCESS_KEY_ID=.*|AWS_ACCESS_KEY_ID=${AWS_ACCESS_KEY_ID}|" ./monitoring-host/.env
                            sed -i "s|^AWS_SECRET_ACCESS_KEY=.*|AWS_SECRET_ACCESS_KEY=${AWS_SECRET_ACCESS_KEY}|" ./monitoring-host/.env
                            
                            # Verify the .env file was created
                            ls -la ./monitoring-host/.env
                        '''
                    }
                }
            }
        }

        stage('Build Docker Image') {
            steps {
                script {
                    docker.build("${DOCKER_IMAGE}:latest", "--file monitoring-host/Dockerfile .")
                }
            }
        }

        stage('Push Docker Image') {
            steps {
                script {
                    docker.withRegistry('https://index.docker.io/v1/', DOCKER_REGISTRY_CREDENTIALS) {
                        docker.image("${DOCKER_IMAGE}:latest").push()
                    }
                }
            }
        }

        stage('Deploy to Swarm') {
            when {
                expression { return env.DEPLOY?.toBoolean() ?: false }
            }
            steps {
                withCredentials([
                    sshUserPrivateKey(credentialsId: 'stag-arnatech-sa-01', keyFileVariable: 'SSH_KEY_FILE'),
                    usernamePassword(credentialsId: 'ard-dockerhub', usernameVariable: 'DOCKERHUB_CREDENTIALS_USR', passwordVariable: 'DOCKERHUB_CREDENTIALS_PSW')
                ]) {
                    sh """
                        # First, copy the .env file to the server
                        scp -i "${SSH_KEY_FILE}" -o StrictHostKeyChecking=no ./monitoring-host/.env root@172.105.124.43:/root/monitoring-host/.env

                        # Then deploy the service
                        ssh -i "${SSH_KEY_FILE}" -o StrictHostKeyChecking=no root@172.105.124.43 -p 22 "
                            
                            # Ensure Docker Swarm is initialized
                            docker swarm init || true

                            # Ensure the overlay network exists
                            docker network create --driver overlay ${NETWORK_NAME} || true

                            # Deploy or update the service
                            docker service rm ${NAMESPACE} || true
                            docker service create --name ${NAMESPACE} \\
                                --replicas ${REPLICAS} \\
                                --network ${NETWORK_NAME} \\
                                --env-file /root/monitoring-host/.env \\
                                ${DOCKER_IMAGE}:latest
                            "
                    """
                }
            }
        }
    }

    post {
        always {
            echo 'Pipeline finished!'
        }
        success {
            echo 'Deployment successful!'
        }
        failure {
            echo 'Pipeline failed.'
        }
    }
} 