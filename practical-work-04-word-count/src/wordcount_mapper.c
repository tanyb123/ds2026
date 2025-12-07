#include "mapreduce.h"
#include <ctype.h>
#include <string.h>

// Word count mapper
// Input: filename (key), line (value)
// Output: word (key), "1" (value)
void wordcount_map(const char *key, const char *value, KeyValueList **output) {
    // This is a simplified version - in real implementation,
    // we'd need to pass the context to emit_intermediate
    // For now, we'll tokenize the line and emit each word
    
    char line[4096];
    strncpy(line, value, sizeof(line) - 1);
    line[sizeof(line) - 1] = '\0';
    
    // Tokenize by whitespace
    char *token = strtok(line, " \t\n\r");
    while (token != NULL) {
        // Convert to lowercase
        char word[256];
        int j = 0;
        for (int i = 0; token[i] != '\0' && j < sizeof(word) - 1; i++) {
            if (isalnum(token[i])) {
                word[j++] = tolower(token[i]);
            }
        }
        word[j] = '\0';
        
        if (j > 0) {
            // Emit (word, "1")
            // Note: In real implementation, we'd use emit_intermediate(ctx, word, "1")
            // For now, we'll add to output list directly
            KeyValueList *new_node = malloc(sizeof(KeyValueList));
            new_node->kv.key = strdup(word);
            new_node->kv.value = strdup("1");
            new_node->next = *output;
            *output = new_node;
        }
        
        token = strtok(NULL, " \t\n\r");
    }
}

