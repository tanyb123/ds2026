#include "mapreduce.h"
#include <stdlib.h>
#include <string.h>

// Word count reducer
// Input: word (key), list of "1" values
// Output: word (key), count (value)
void wordcount_reduce(const char *key, const char **values, int num_values, FILE *output) {
    // Count the number of values (each is "1")
    int count = num_values;
    
    // Emit final result
    char count_str[32];
    snprintf(count_str, sizeof(count_str), "%d", count);
    emit_final(output, key, count_str);
}

