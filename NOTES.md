# NOTES.md

## Goroutine Topology

-- Below menitoneed Data FLow is the main Data Flow of the GO Routine

         Files → fileCh → Workers → resultCh → Printer → Output

main
├── Signal goroutine (Cancel) listens sigCh → calls cancel() on Ctrl+C
├── Walker goroutine ( Identify File Paths ) walkDir() → writes fileCh → closes fileCh on exit
├── Worker Goroutine worker × N → reads fileCh → processFile() → writes resultCh
└── Printer goroutine reads resultCh → stdout → closes printerDone

- fileCh — closed by the **Walker** (`defer close(fileCh)`)
- resultCh — closed by **Main** after `workerWg.Wait()`
- printerDone — closed by **Printer** when resultCh is drained

**Normal exit:** Walker Chanel finishes → fileCh closes → workers exit range loop → workerWg unblocks → main closes resultCh → printer closes printerDone → main prints summary.

**Cancellation:** ctx.Done() fires → walker and workers return early → same shutdown sequence follows.

## Goroutine Leak Prevention

If `resultCh` is never closed, the printer blocks forever on `range resultCh` and `<-printerDone` never unblocks — leaking both goroutines.

Prevented by:

```go
workerWg.Wait()  // all workers done, no more sends possible
close(resultCh)  // safe to close now
```

Closing before `workerWg.Wait()` would panic and be send on a closed chanel.

## Why contenxt.Context (ctx) Is Checked Inside the Scan Loop

A file can have thousands of lines. so if you ony check the cancellation func after each file a large file can take a large time to stop, if you check the method ctx.Done() on every line within every file, then it stops instantly after the next line rather than just reading the rest of the whole file

## Why Workers Must Not Close 'resultCh'

Here Likely, if a working worker is closed if others are workers are still running this will cause a mishap and will result in unnecessary damage and harm. here only in main after the method workerWg.wait(), knows that all the wokers that are assigned are finished is the best time to close all the workers
`

## One Thing I Changed from the AI Output

Here, mainly after i developed the program the the bash command wouldnt accept the directory in the form ./ or .\ only after using the syntax "=" it worked because it identified the rest as another arg, so with the help of AI we implemeneted a rebuildArgs to glue both args accepting the bash command into a single command.
