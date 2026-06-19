#include <stdio.h>
#include <stdlib.h>

#include "singboxffi.h"

static void print_owned_string(const char *label, char *value) {
    if (value == NULL) {
        return;
    }
    printf("%s%s\n", label, value);
    sb_free_string(value);
}

static int fail_with_error(const char *step, char *err) {
    if (err != NULL) {
        fprintf(stderr, "%s failed: %s\n", step, err);
        sb_free_string(err);
    } else {
        fprintf(stderr, "%s failed\n", step);
    }
    return 1;
}

int main(void) {
    char *version = (char *)sb_version();
    char *go_version = (char *)sb_go_version();
    char *err = NULL;

    print_owned_string("sing-box version: ", version);
    print_owned_string("go version: ", go_version);

    sb_init_options opts = {
        .base_path = ".",
        .working_path = ".",
        .temp_path = ".",
        .locale = NULL,
        .command_secret = "example-secret",
        .command_port = 0,
        .log_max_lines = 300,
        .debug = false,
        .oom_killer_enabled = false,
        .oom_killer_disabled = true,
        .oom_memory_limit = 0,
    };

    if (sb_init(&opts, &err) != 0) {
        return fail_with_error("sb_init", err);
    }

    char config_json[] =
        "{"
        "  \"log\": {\"level\": \"info\"},"
        "  \"inbounds\": ["
        "    {"
        "      \"type\": \"mixed\","
        "      \"tag\": \"mixed-in\","
        "      \"listen\": \"127.0.0.1\","
        "      \"listen_port\": 2080"
        "    }"
        "  ],"
        "  \"outbounds\": ["
        "    {\"type\": \"direct\", \"tag\": \"direct\"}"
        "  ],"
        "  \"route\": {\"final\": \"direct\"}"
        "}";

    if (sb_check_config(config_json, &err) != 0) {
        return fail_with_error("sb_check_config", err);
    }

    puts("config is valid");
    return 0;
}
