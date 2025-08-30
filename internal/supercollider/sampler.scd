(
s.waitForBoot({
	Routine{
		~serverLatency = 0.1;
		~synthPlayback = nil;
		~synthRecord = Dictionary.new();
		~samplesPlaying = Dictionary.new();
		~synthsPlaying = Dictionary.new();



		SynthDef("PolyPerc",{
			arg trackId,
			trackVolume = 0,
			pan=0,
			note=0,
			gate=1,
			attack=0.1,
			decay=1.0,
			sustain=0.7,
			release=0.5,
			effectDryOut,
			effectReverb = 0.0, effectReverbOut,
			effectComb = 0.0, effectCombOut,
			a=0,b=0,c=0,d=0 // a,b,c,d vary between 0 and 1.0
			;
			var snd;
			var env;
			snd = Pulse.ar(note.midicps, 0.5);
			snd = MoogFF.ar(snd, 2000, 1.5);
			env=EnvGen.ar(Env.adsr(attack,decay,sustain,release),gate,doneAction:2);
			snd = Pan2.ar(snd,pan);
			Out.ar(effectDryOut, snd);
			Out.ar(effectCombOut, snd*effectComb);
			Out.ar(effectReverbOut, snd*effectReverb);
		}).add;


		SynthDef("synthPad", {
			arg trackId,
			trackVolume = 0,
			pan=0,
			note=0,
			attack=1.0,
			effectDryOut,
			effectReverb = 0.0, effectReverbOut,
			effectComb = 0.0, effectCombOut,
			a=0,b=0,c=0,d=0 // a,b,c,d vary between 0 and 1.0
			;
			var snd;
			var freq = note.midicps;
			snd = SawDPW.ar(freq * ((-3..2) * 0.05).midiratio * [1, 2, 1, 4, 1, 2]);
			snd = Splay.ar(snd);
			snd = MoogFF.ar(snd, XLine.kr(100,rrand(6000,\lpf.kr(18000)),attack*(1/\swell.kr(1))), 0);
			snd = snd * EnvGen.ar(Env.asr(attack, 1.0, \release.kr(0.5)),\gate.kr(1) * ToggleFF.kr(1-TDelay.kr(DC.kr(1),\duration.kr(1))),doneAction:2);
			snd = Balance2.ar(snd[0], snd[1], pan);
			snd = snd * -10.dbamp * trackVolume.dbamp;
			Out.ar(effectDryOut, snd);
			Out.ar(effectCombOut, snd*effectComb);
			Out.ar(effectReverbOut, snd*effectReverb);
		}).add;

		SynthDef("diskout", { arg bufnum=0, inbus=0, gate=1;
			var snd = In.ar(inbus,2);
			snd = snd * EnvGen.ar(Env.adsr(0.001,0.0,1.0,1.0),gate,doneAction:2);
			DiskOut.ar(bufnum,snd);
		}).add;

		2.do({
			arg ch;
			SynthDef("playback"++(ch+1),{
				arg out,buf,gate = 1;
				var env = EnvGen.ar(Env.adsr(0.01,0.0,1.0,0.2),gate,doneAction:2);
				var snd = PlayBuf.ar(ch+1, buf, BufRateScale.kr(buf), loop: 0, doneAction:2);
				if (ch<1,{
					snd = Pan2.ar(snd,0);
				});
				Out.ar(out, (snd*env));
			}).add;
		});

		2.do({
			arg ch;
			SynthDef("sampler"++(ch+1),{
				arg buf,
				trackId=0,
				volumeDB=0,
				rate = 1.0,
				pitch = 0.0,
				gate = 1,
				t_trig = 1,
				xfade=0.01,
				bpmSource=120,
				bpmTarget=120,
				lowPassFilter=20000,
				highPassFilter=20,
				trackVolume = 0.0,
				pan=0,
				// retriggering
				retrigNumTotal=0,
				retrigRateChangeBeats=1,
				retrigRateStart=1,
				retrigRateEnd=0,
				retrigPitchChange=0,
				retrigVolumeChange=0,
				sliceAttackBeats = 0.001,
				sliceDurationBeats = 0,
				sliceReleaseBeats = 0.001,
				sliceNum = 0,
				sliceCount = 32, // number of slices to cut the sample into
				effectDry = 1.0,
				effectDryOut,
				effectComb = 0.0,
				effectCombOut,
				effectDelay = 0.0,
				effectReverse = 0,
				effectTimestretchStart = 0,
				effectTimestretchEnd = 0,
				effectTimestretchBeats = 0,
				effectReverb = 0.0,
				effectLPFStart = 20000,
				effectLPFEnd = 0,
				effectLPFBeats = 0,
				effectReverbOut
				;
				var snd,sndA,sndB,crossfade,aOrB;
				var sndDry, sndReverb, sndComb;
				var posA,posB;
				var retrigTrig, retrigCount;
				var frames=BufFrames.ir(buf);
				var seconds=BufDur.ir(buf)*bpmSource/bpmTarget;
				var sliceSeconds = seconds / sliceCount;
				var pos = 0.0;
				var sliceTrigger = t_trig;
				var beatDuration = 60 / bpmTarget;
				var retrigCountFeedback = 0;
				var timestretchPos, timestretchRate, effectTimestretch;

				sliceNum = sliceNum.mod(sliceCount);
				// convert slice beat to seconds
				pos = Select.kr(effectReverse>0,[
					sliceNum / sliceCount,
					((sliceNum+1).mod(sliceCount))/sliceCount
				]);

				// set effectLPFEnd to effectLPFStart if it is 0
				effectLPFEnd = Select.kr(effectLPFEnd < 0.001, [
					effectLPFEnd,
					effectLPFStart,
				]);
				// set effectTimestretch to beginning if it is 0
				effectTimestretchEnd = Select.kr(effectTimestretchEnd< 0.001,[
					effectTimestretchEnd,
					effectTimestretchStart,
				]);
				effectTimestretch = Line.kr(effectTimestretchStart, effectTimestretchEnd, effectTimestretchBeats*beatDuration);
				timestretchRate = Select.ar(effectTimestretch>1,[
					DC.ar(pos*frames),
					Phasor.ar(trig:TDelay.ar(Impulse.ar(0, 0),0),rate:rate/effectTimestretch,start:0,end:frames,resetPos:(pos*frames)).floor,
				]);


				// make sure retrigRateEnd is at least retrigRateStart if its 0
				retrigRateEnd = Select.kr(retrigRateEnd < 0.001, [retrigRateEnd, retrigRateStart]);

				// if sliceDurationBeats = 0, make it infinite
				sliceDurationBeats = Select.kr(sliceDurationBeats < 0.001, [sliceDurationBeats, (seconds/sliceCount)/(60/bpmSource)]);

				// Calculate rate
				rate = rate*BufRateScale.ir(buf)*bpmTarget/bpmSource;
				// Calculate pitch
				pitch = pitch + (retrigNumTotal * retrigPitchChange.neg);
				rate = rate * (2 ** (pitch/12.0));

				// Calculate retriggers
				retrigCountFeedback = LocalIn.ar(1);
				retrigTrig = Impulse.ar(XLine.ar(retrigRateStart,retrigRateEnd,beatDuration*retrigRateChangeBeats) / beatDuration);
				retrigCount = PulseCount.ar(retrigTrig * (retrigCountFeedback<retrigNumTotal));
				sliceTrigger = sliceTrigger + (retrigTrig*(retrigCount<retrigNumTotal));

				// Increase pitch based on retrigger
				// rate = rate * (2 ** (retrigCount / 12));
				// Lower pitch based on retrigger
				rate = rate * (2 ** (retrigPitchChange * (retrigCount / 12)));
				rate = rate.abs;
				// Do reverse
				rate = Select.kr(effectReverse>0,[
					rate,
					rate * -1,
				]);

				// Change volume based on retrigger
				volumeDB = volumeDB + (A2K.kr(retrigCount) * retrigVolumeChange) + (retrigVolumeChange.neg * retrigNumTotal * (retrigVolumeChange>0));


				// do timestretching
				sliceTrigger = sliceTrigger + ((effectTimestretch>1)*Impulse.ar(effectTimestretch/sliceSeconds));


				// Determine whether to toggle playback
				aOrB=ToggleFF.ar(sliceTrigger);
				crossfade=VarLag.ar(K2A.ar(aOrB),xfade,warp:\sine);

				posA=Phasor.ar(
					trig:(1-aOrB),
					rate:rate,
					end:frames,
					resetPos:timestretchRate,
				);
				posB=Phasor.ar(
					trig:aOrB,
					rate:rate,
					end:frames,
					resetPos:timestretchRate,
				);
				snd=(BufRd.ar(
					numChannels:ch+1,
					bufnum:buf,
					phase:posA,
				)*crossfade)+(BufRd.ar(
					numChannels:ch+1,
					bufnum:buf,
					phase:posB,
				)*(1-crossfade));

				snd = snd * Lag.kr(volumeDB.dbamp,0.2);

				// full envelope
				snd = snd * EnvGen.ar(Env.adsr(xfade,0.0,1.0,xfade,curve:\sine),gate,doneAction:2);

				// slice envelope
				snd = snd * EnvGen.ar(Env.new([0,1,1,0],[sliceAttackBeats,sliceDurationBeats,sliceReleaseBeats]*beatDuration,[-4,4]), sliceTrigger);


				// filtering
				snd = RLPF.ar(snd, XLine.kr(effectLPFStart, effectLPFEnd, effectLPFBeats*beatDuration), 0.303);
				DetectSilence.ar(snd,time:1,doneAction:2);

				// filtering
				snd = RLPF.ar(snd, lowPassFilter, 0.707);
				snd = RHPF.ar(snd, highPassFilter, 0.707);

				// panning
				if (ch<1,{
					snd = Pan2.ar(snd,pan);
				},{
					snd = Balance2.ar(snd[0],snd[1],pan);
				});

				// volume
				snd = snd * trackVolume.dbamp;

				// Feedback
				LocalOut.ar(retrigCount);
				Out.ar(effectDryOut, snd*effectDry);
				Out.ar(effectCombOut, snd*XLine.kr(0.001,effectComb,sliceDurationBeats*beatDuration/2));
				Out.ar(effectReverbOut, snd*XLine.kr(0.001,effectReverb,sliceDurationBeats*beatDuration/2));
			}).add;
		});

		SynthDef("out",{
			arg busReverb, busDry, busComb, busDisk,
			volumeDB=0.0,
			reverbAmt=0.1,
			pregain=0.0,
			postgain=0.0,
			bias=6.0.neg,
			saturation=6.0.neg,
			drive=6.0.neg,
			shimmer=1.0,
			combAmt=0.0;
			var sndWet = In.ar(busReverb,2);
			var sndDry = In.ar(busDry,2);
			var sndComb = In.ar(busComb,2);
			var snd = 				sndDry;
			SendReply.kr(Impulse.kr(30),'/track_volume',[0,Lag.kr(Amplitude.kr(snd[0],0.3,0.3).ampdb,3)]);

			// add in comb
			snd = snd + ((0.5*sndComb)+
				Pan2.ar(CombN.ar(sndComb[0], 0.02, Rand(0.01,0.06), Rand(1,2)),-1,0.9) +
				Pan2.ar(CombN.ar(sndComb[1], 0.02, Rand(0.01,0.06), Rand(1,2)),1,0.9));

			// add in reverb
			sndWet = DelayN.ar(sndWet, 0.03, 0.03);
			sndWet = sndWet + PitchShift.ar(sndWet, 0.13, 2,0,1,1*shimmer/2);
			sndWet = sndWet + PitchShift.ar(sndWet, 0.1, 4,0,1,0.5*shimmer/2);
			sndWet = sndWet + PitchShift.ar(sndWet, 0.1, 8,0,1,0.125*shimmer/2);
			snd = snd + Fverb.ar(sndWet[0],sndWet[1],200,
				tail_density: LFNoise2.ar(1/3).range(50,90),
				decay: LFNoise2.ar(1/3).range(50,70)
			);

			snd = RHPF.ar(snd,60,0.303);
			snd = snd * Lag.kr(pregain).dbamp;
			snd = AnalogTape.ar(snd,
				bias: Lag.kr(bias).dbamp,
				saturation: Lag.kr(saturation).dbamp,
				drive: Lag.kr(drive).dbamp,
				oversample: 2,
			);
			snd = snd * Lag.kr(volumeDB).dbamp * Lag.kr(postgain).dbamp;

			SendReply.kr(Impulse.kr(30),'/waveform',Normalizer.ar(LPF.ar(snd,3000)));
			ReplaceOut.ar(0,snd);
			Out.ar(busDisk, snd);
		}).add;

		s.sync;
		~busDry = Bus.audio(s, 2);
		~busReverb = Bus.audio(s, 2);
		~busComb = Bus.audio(s, 2);
		~busDisk = Bus.audio(s, 2);
		s.sync;
		~synOut = Synth.tail(s,"out",[
			busReverb: ~busReverb,
			busDry: ~busDry,
			busComb: ~busComb,
			busDisk: ~busDisk,
			volumeDB: -24,
		]);
		s.sync;
		~sampleCache = Dictionary.new();


		~playSynthFromMsg = {
			arg msg;
			var synName = 1000000.rand.asString;
			var track = msg[1].asInteger;
			var argLast;
			var noteOn = msg[2].asInteger;
			var args = [
				\effectDryOut, ~busDry,
				\effectCombOut, ~busComb,
				\effectReverbOut, ~busReverb,
				\trackId, track,
			];
			["track",track,"noteOn",noteOn].postln;
			// create a new dictionary if it doesn't exist
			if (~synthsPlaying.at(track).isNil,{
				~synthsPlaying.put(track, Dictionary.new());
			});
			// stop all currently playing synths for this track
			~synthsPlaying.at(track).values.do({
				arg syn;
				if (syn.isPlaying,{
					if (syn.notNil,{
						syn.set(\gate,0);
					});
				});
			});
			if (noteOn > 0,{
				args = args ++ msg[3..];
				~synthsPlaying.at(track).put(synName,
					Synth.head(s,"PolyPerc",args).onFree({
						[synName,"freed"].postln;
						~synthsPlaying.at(track).removeAt(synName);
					});
				);
				NodeWatcher.register(~synthsPlaying.at(track).at(synName));
			});
		};


		~playFromMsg = {
			arg msg, b;
			var synName = 1000000.rand.asString;
			var track = msg[2].asInteger;
			var argLast;
			var args = [
				\buf,b,
				\effectDryOut, ~busDry,
				\effectCombOut, ~busComb,
				\effectReverbOut, ~busReverb,
				\trackId, track,
			];
			// create a new dictionary for the track if it doesn't exist
			if (~samplesPlaying.at(track).isNil,{
				~samplesPlaying.put(track, Dictionary.new());
			});
			// stop all currently playing synths for this track
			~samplesPlaying.at(track).values.do({
				arg syn;
				if (syn.isPlaying,{
					if (syn.notNil,{
						syn.set(\gate,0);
					});
				});
			});
			// add msg[2:] array to args
			args = args ++ msg[3..];
			args.do({
				arg v,i;
				if (v.isNumber,{
					args.put(i,v.round(1.0/16.0));
					[argLast,v.round(1.0/16.0)].postln;
				},{
					argLast = v;
				});
			});


			~samplesPlaying.at(track).put(synName,
				Synth.head(s,"sampler"++(b.numChannels),args).onFree({
					[b,"freed"].postln;
					~samplesPlaying.at(track).removeAt(synName);
				});
			);

		};

		OSCFunc({ |msg|
			var filename = msg[1];
			msg.postln;
			if (~sampleCache.at(filename).isNil,{
				// load sample
				~sampleCache.put(filename, Buffer.read(s,filename,action:{ |b|
					["loaded",b,b.numChannels].postln;
					~playFromMsg.(msg,b);
				}));
			},{
				~playFromMsg.(msg,~sampleCache.at(filename));
			});
		},'/sampler');
		OSCFunc({ |msg|
		msg.postln;
			~playSynthFromMsg.(msg);
		},'/instrument');

		OSCFunc({ |msg|
			// stop all currently playing synths in all tracks
			~samplesPlaying.values.do({
				arg track;
				track.values.do({ arg syn;
					if (syn.isPlaying,{
						if (syn.notNil,{
							syn.set(\gate,0);
						});
					});
				});
			});
			// stop playback synth if it exists
			if (~synthPlayback.notNil,{
				if (~synthPlayback.isPlaying,{
					[~synthPlayback,"stopped"].postln;
					~synthPlayback.set(\gate,0);
				});
			});
		},'/stop');
		OSCFunc({ |msg|
			NetAddr.new("127.0.0.1", 57121).sendMsg("/waveform", msg[3]);
		},'/waveform');
		OSCFunc({ |msg|
			NetAddr.new("127.0.0.1", 57121).sendMsg("/track_volume", msg[3],msg[4]);
		},'/track_volume');
		OSCFunc({ |msg|
			var filename = msg[1].asString;
			var doRecord = msg[2].asInteger;
			var synRecord = ~synthRecord.at(filename);
			// stop recording if recording
			if (synRecord.notNil,{
				if (synRecord.isPlaying,{
					[synRecord,"stopped"].postln;
					synRecord.set(\gate,0);
					~synthRecord.removeAt(filename);
					["stopping recording",filename].postln;
				});
			});
			if (doRecord<1,{
				// do nothing, recording already released
			}, {
				// start a recording
				var recordingBuffer=Buffer.alloc(Server.default,65536,2);
				var pathname=filename;
				["starting recording",filename].postln;
				recordingBuffer.write(pathname.standardizePath,PathName.new(pathname.standardizePath).extension,"int16",0,0,true);
				~synthRecord.put(filename,Synth.tail(s,"diskout",[
					\bufnum,recordingBuffer.bufnum,
					\inbus,~busDisk,
					\gate,1,
				]).onFree({
					[recordingBuffer,"freed"].postln;
					recordingBuffer.free;
				}));
				NodeWatcher.register(~synthRecord.at(filename));
			});
		},'/record');
		OSCFunc({ |msg|
			var filename = msg[1];
			var gate = msg[2].asInteger;
			msg.postln;
			if (~synthPlayback.notNil,{
				if (~synthPlayback.isPlaying,{
					[~synthPlayback,"stopped"].postln;
					~synthPlayback.set(\gate,0);
				});
			});
			if (gate>0,{
				Buffer.read(s,filename,action:{ |b|
					~synthPlayback = Synth.head(s,"playback"++(b.numChannels),[
						\buf,b,
						\out,~busDry,
						\gate,1,
					]).onFree({
						[b,"freed"].postln;
						b.free;
					});
					NodeWatcher.register(~synthPlayback);
				});
			});
		},'/playback');
		OSCFunc({ |msg|
			["setting",msg[1],msg[2]].postln;
			~synOut.set(msg[1],msg[2]);
		},'/set');

		["loaded",NetAddr.langPort, NetAddr.localAddr].postln;

		s.sync;
		Routine {
			inf.do({
				NetAddr.new("127.0.0.1", 57121).sendMsg("/cpuusage", s.avgCPU);
				1.sleep;
			});
		}.play;

		// ~bpmTarget = 180;
		// ~sliceNum = 32;
		// Routine {
		// 	inf.do({ arg beat;
		// 		var sliceNum = beat.mod(~sliceNum);
		// 		["playing",beat].postln;
		// 		NetAddr.new("127.0.0.1", NetAddr.langPort).sendMsg("/sampler", "amen_5c2d11c8_beats16_bpm170.flac",
		// 			\sliceCount, 32,
		// 			\sliceNum, sliceNum,
		// 			\sliceDurationBeats, 100,
		// 			\bpmSource, 170,
		// 			\bpmTarget, ~bpmTarget,
		// 			\effectTimestretchStart, 3.rand,
		// 			\effectTimestretchEnd, 8.rand,
		// 			\effectTimestretchBeats, 8.rand,
		// 			\retrigNumTotal, 16.rand,
		// 			\retrigRateStart, 16.rand,
		// 			\retrigPitchChange,-1,
		// 			\retrigVolumeChange,-2,
		// 			\effectComb, (100.rand)/100.0,
		// 		);

		// 		(60/~bpmTarget/2).wait;
		// 	});
		// }.play;
		// Routine {
		// 	NetAddr.new("127.0.0.1", NetAddr.langPort).sendMsg("/note_on");
		// }.play;
	}.play;
});
)
