# from pydub import AudioSegment
# import os

# def split_audio(file_path, output_dir, segment_duration=3*60*1000):
#     """
#     Dzieli plik audio na fragmenty o określonej długości.
    
#     :param file_path: Ścieżka do pliku audio
#     :param output_dir: Katalog, w którym zostaną zapisane fragmenty
#     :param segment_duration: Długość każdego fragmentu w milisekundach (domyślnie 10 minut)
#     """
#     # Ładowanie pliku audio
#     try:
#         audio = AudioSegment.from_file(file_path)
#     except Exception as e:
#         print(f"Błąd podczas ładowania pliku audio: {e}")
#         return

#     # Ustalanie liczby fragmentów
#     total_duration = len(audio)
#     num_segments = (total_duration + segment_duration - 1) // segment_duration

#     # Tworzenie katalogu wyjściowego, jeśli nie istnieje
#     if not os.path.exists(output_dir):
#         os.makedirs(output_dir)

#     # Dzielenie pliku audio
#     for i in range(num_segments):
#         start_time = i * segment_duration
#         end_time = min(start_time + segment_duration, total_duration)
#         segment = audio[start_time:end_time]
        
#         # Zapis fragmentu do pliku
#         output_file = os.path.join(output_dir, f"segment_{i+1}.wav")
#         segment.export(output_file, format="wav")
#         print(f"Zapisano fragment: {output_file}")

#     print("Podział pliku audio zakończony.")

# # Użycie skryptu
# file_path = "PATH TO FILE WHICH YOU WANT TO DIVIDE INTO SMALLER PARTS"
# output_dir = "/segments" 
# split_audio(file_path, output_dir)
