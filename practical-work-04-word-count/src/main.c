#include <stdio.h>
#include <stdlib.h>
#include "mapreduce.h"
#include "wordcount_mapper.c"
#include "wordcount_reducer.c"

int main(int argc, char **argv) {
    if (argc < 3) {
        fprintf(stderr, "Usage: %s [--mappers N] [--reducers M] <input_dir> <output_dir>\n", argv[0]);
        return 1;
    }
    
    int num_mappers = 2;
    int num_reducers = 1;
    const char *input_dir = NULL;
    const char *output_dir = NULL;
    
    // Parse arguments
    for (int i = 1; i < argc; i++) {
        if (strcmp(argv[i], "--mappers") == 0 && i + 1 < argc) {
            num_mappers = atoi(argv[++i]);
        } else if (strcmp(argv[i], "--reducers") == 0 && i + 1 < argc) {
            num_reducers = atoi(argv[++i]);
        } else if (input_dir == NULL) {
            input_dir = argv[i];
        } else if (output_dir == NULL) {
            output_dir = argv[i];
        }
    }
    
    if (input_dir == NULL || output_dir == NULL) {
        fprintf(stderr, "Error: input_dir and output_dir are required\n");
        return 1;
    }
    
    printf("Word Count MapReduce\n");
    printf("====================\n");
    printf("Input directory: %s\n", input_dir);
    printf("Output directory: %s\n", output_dir);
    printf("Number of mappers: %d\n", num_mappers);
    printf("Number of reducers: %d\n", num_reducers);
    printf("\n");
    
    // Initialize MapReduce context
    MapReduceContext ctx;
    if (mapreduce_init(&ctx, num_mappers, num_reducers, input_dir, output_dir,
                       wordcount_map, wordcount_reduce) != 0) {
        fprintf(stderr, "Error: Failed to initialize MapReduce\n");
        return 1;
    }
    
    // Run MapReduce job
    if (mapreduce_run(&ctx) != 0) {
        fprintf(stderr, "Error: MapReduce job failed\n");
        mapreduce_cleanup(&ctx);
        return 1;
    }
    
    // Cleanup
    mapreduce_cleanup(&ctx);
    
    printf("\nWord count completed successfully!\n");
    printf("Check output files in %s/\n", output_dir);
    
    return 0;
}

