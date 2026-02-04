FROM rabbitmq:4.2-management-alpine

# Download the plugin
RUN wget https://github.com/rabbitmq/rabbitmq-delayed-message-exchange/releases/download/v4.2.0/rabbitmq_delayed_message_exchange-4.2.0.ez && \
    mv rabbitmq_delayed_message_exchange-4.2.0.ez /plugins/

# Enable the plugin
RUN rabbitmq-plugins enable --offline rabbitmq_delayed_message_exchange
